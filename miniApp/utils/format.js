function pad(value) {
  return String(value).padStart(2, "0");
}

function dateText(value) {
  const date = value ? new Date(value) : new Date();
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`;
}

function paginate(list, page, pageSize) {
  const current = Math.max(1, Number(page || 1));
  const size = Math.max(1, Number(pageSize || 20));
  const start = (current - 1) * size;
  return {
    page: current,
    pageSize: size,
    total: list.length,
    list: list.slice(start, start + size),
  };
}

module.exports = { dateText, paginate };
