const api = require("../../services/api");
const { go } = require("../../utils/navigation");

const WEEK_LABELS = ["日", "一", "二", "三", "四", "五", "六"];
const WEEKDAY_NAMES = ["周日", "周一", "周二", "周三", "周四", "周五", "周六"];

function pad(value) {
  return String(value).padStart(2, "0");
}

function toDateKey(date) {
  return [date.getFullYear(), pad(date.getMonth() + 1), pad(date.getDate())].join("-");
}

function parseDateKey(key) {
  const parts = String(key || "").split("-").map(Number);
  return new Date(parts[0], parts[1] - 1, parts[2] || 1);
}

function toMonthCursor(date) {
  return [date.getFullYear(), pad(date.getMonth() + 1)].join("-");
}

function formatMonthTitle(cursor) {
  const parts = cursor.split("-").map(Number);
  return parts[0] + "年" + parts[1] + "月";
}

function formatDateTitle(key) {
  const date = parseDateKey(key);
  return (date.getMonth() + 1) + "月" + date.getDate() + "日";
}

function formatFormDate(key) {
  const date = parseDateKey(key);
  return formatDateTitle(key) + " · " + WEEKDAY_NAMES[date.getDay()];
}

function formatClock(value, fallbackIndex) {
  if (value) {
    const date = new Date(value);
    if (!Number.isNaN(date.getTime())) return pad(date.getHours()) + ":" + pad(date.getMinutes());
  }
  return fallbackIndex % 2 === 0 ? "18:42" : "12:16";
}

function normalizeCheckin(item, index) {
  const checkedAt = item.checkedAt || "";
  return {
    ...item,
    date: checkedAt.slice(0, 10),
    note: item.note || "",
    displayTime: formatClock(checkedAt, index),
  };
}

function buildMonthView(cursor, checkins, selectedDate) {
  const parts = cursor.split("-").map(Number);
  const year = parts[0];
  const month = parts[1];
  const prefix = cursor + "-";
  const monthlyCheckins = checkins.filter((item) => String(item.date || "").startsWith(prefix));
  const checkedDates = new Set(monthlyCheckins.map((item) => item.date));
  const todayKey = toDateKey(new Date());
  const fallbackDate = monthlyCheckins.length
    ? monthlyCheckins[0].date
    : (todayKey.startsWith(prefix) ? todayKey : prefix + "01");
  const activeDate = selectedDate && selectedDate.startsWith(prefix) ? selectedDate : fallbackDate;
  const firstWeekday = new Date(year, month - 1, 1).getDay();
  const dayCount = new Date(year, month, 0).getDate();
  const blanks = Array.from({ length: firstWeekday }, (_, index) => ({
    key: "blank-" + index,
    day: 0,
    date: "",
  }));
  const days = Array.from({ length: dayCount }, (_, index) => {
    const day = index + 1;
    const date = prefix + pad(day);
    return {
      key: date,
      day,
      date,
      checked: checkedDates.has(date),
      selected: date === activeDate,
      today: date === todayKey,
    };
  });
  const selectedCheckins = monthlyCheckins
    .filter((item) => item.date === activeDate)
    .map(normalizeCheckin);
  const selectedRewardClaimed = selectedCheckins.some((item) => item.rewarded);

  return {
    monthCursor: cursor,
    displayMonth: formatMonthTitle(cursor),
    monthCheckinCount: monthlyCheckins.length,
    calendar: [...blanks, ...days],
    selectedDate: activeDate,
    selectedDateLabel: formatDateTitle(activeDate),
    selectedCheckins,
    selectedRewardClaimed,
    rewardMessage: activeDate === todayKey ? "今天的首次奖励已领取" : "该日首次打卡奖励已领取",
  };
}

Page({
  data: {
    weekLabels: WEEK_LABELS,
    allCheckins: [],
    dishes: [],
    totalCheckinDays: 0,
    monthCursor: "",
    displayMonth: "",
    monthCheckinCount: 0,
    calendar: [],
    selectedDate: "",
    selectedDateLabel: "",
    selectedCheckins: [],
    selectedRewardClaimed: false,
    rewardMessage: "",
    showAdd: false,
    showSuccess: false,
    formDateKey: "",
    formDateLabel: "",
    photo: "",
    dishName: "",
    note: "",
    canSubmit: false,
    submitting: false,
    successCheckin: null,
    successRewardPoints: 0,
    successDateLabel: "",
  },

  async onLoad() {
    await this.load();
  },

  async load(preferredDate) {
    const [checkinPage, dishesPage, bootstrap] = await Promise.all([
      api.listCheckins({ page: 1, pageSize: 100 }),
      api.listDishes({ page: 1, pageSize: 100, status: "usable" }),
      api.bootstrap(),
    ]);
    const allCheckins = (checkinPage.list || []).map(normalizeCheckin);
    const selectedDate = preferredDate || (allCheckins[0] && allCheckins[0].date) || toDateKey(new Date());
    const monthCursor = selectedDate.slice(0, 7);
    this.setData({
      allCheckins,
      dishes: dishesPage.list || [],
      totalCheckinDays: (bootstrap.profile && bootstrap.profile.checkinDays) || checkinPage.total || 0,
      ...buildMonthView(monthCursor, allCheckins, selectedDate),
    });
  },

  changeMonth(event) {
    const offset = Number(event.currentTarget.dataset.offset || 0);
    const parts = this.data.monthCursor.split("-").map(Number);
    const next = new Date(parts[0], parts[1] - 1 + offset, 1);
    const cursor = toMonthCursor(next);
    this.setData(buildMonthView(cursor, this.data.allCheckins, ""));
  },

  selectDate(event) {
    const date = event.currentTarget.dataset.date;
    if (!date) return;
    this.setData(buildMonthView(this.data.monthCursor, this.data.allCheckins, date));
  },

  openDish(event) {
    const id = event.currentTarget.dataset.id;
    if (id) go("/pages/dish/detail", { id });
  },

  openAdd() {
    const formDateKey = toDateKey(new Date());
    this.setData({
      showAdd: true,
      formDateKey,
      formDateLabel: formatFormDate(formDateKey),
      photo: "",
      dishName: "",
      note: "",
      canSubmit: false,
      submitting: false,
    });
  },

  closeAdd() {
    if (this.data.submitting) return;
    this.setData({ showAdd: false });
  },

  noop() {},

  choosePhoto() {
    wx.chooseMedia({
      count: 1,
      mediaType: ["image"],
      sourceType: ["album", "camera"],
      success: (result) => {
        const photo = result.tempFiles[0].tempFilePath;
        const detected = this.data.dishes.find((dish) => dish.name === "番茄炒蛋") || this.data.dishes[0];
        const dishName = this.data.dishName || (detected ? detected.name : "");
        this.setData({
          photo,
          dishName,
          canSubmit: Boolean(photo && dishName.trim()),
        });
      },
    });
  },

  removePhoto() {
    this.setData({
      photo: "",
      dishName: "",
      canSubmit: false,
    });
  },

  inputDishName(event) {
    const dishName = event.detail.value;
    this.setData({
      dishName,
      canSubmit: Boolean(this.data.photo && dishName.trim()),
    });
  },

  inputNote(event) {
    this.setData({ note: event.detail.value });
  },

  async submit() {
    if (!this.data.canSubmit || this.data.submitting) return;
    const dishName = this.data.dishName.trim();
    this.setData({ submitting: true });
    try {
      const result = await api.createCheckin({
        dishName,
        image: this.data.photo,
        note: this.data.note.trim(),
      });
      const successCheckin = normalizeCheckin({
        ...result.checkin,
        dishName,
        note: this.data.note.trim(),
        imageUrl: result.checkin.imageUrl || this.data.photo,
      }, 0);
      const checkedDate = result.checkin.checkedAt.slice(0, 10);
      await this.load(checkedDate);
      this.setData({
        showAdd: false,
        showSuccess: true,
        successCheckin,
        successRewardPoints: result.rewardAmount || 0,
        successDateLabel: formatDateTitle(checkedDate) + " " + successCheckin.displayTime,
      });
    } catch (error) {
      wx.showToast({ title: error.message || "打卡失败，请重试", icon: "none" });
    } finally {
      this.setData({ submitting: false });
    }
  },

  addAsDish() {
    const name = this.data.successCheckin ? this.data.successCheckin.dishName : "";
    this.setData({ showSuccess: false });
    go("/pages/dish/add-entry", { name });
  },

  finishSuccess() {
    this.setData({ showSuccess: false });
  },
});
