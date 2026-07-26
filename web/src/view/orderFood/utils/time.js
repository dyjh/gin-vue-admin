// ORDER_FOOD_TIME_ZONE 是点餐业务管理端统一使用的时区。
export const ORDER_FOOD_TIME_ZONE = 'Asia/Shanghai'

const SHANGHAI_OFFSET = '+08:00'
const LOCAL_DATE_TIME_PATTERN = /^(\d{4})-(\d{2})-(\d{2})[ T](\d{2}):(\d{2}):(\d{2})$/
const zonedDateTimeFormatter = new Intl.DateTimeFormat('zh-CN', {
  timeZone: ORDER_FOOD_TIME_ZONE,
  year: 'numeric',
  month: '2-digit',
  day: '2-digit',
  hour: '2-digit',
  minute: '2-digit',
  second: '2-digit',
  hourCycle: 'h23'
})

// formatOrderFoodDateTime 将接口时间统一回显为上海时区的年月日时分秒。
export const formatOrderFoodDateTime = (value) => {
  if (value === null || value === '' || typeof value === 'undefined') return ''
  const date = value instanceof Date ? value : new Date(value)
  if (Number.isNaN(date.getTime())) return ''

  const parts = Object.fromEntries(
    zonedDateTimeFormatter
      .formatToParts(date)
      .filter((part) => part.type !== 'literal')
      .map((part) => [part.type, part.value])
  )
  return `${parts.year}-${parts.month}-${parts.day} ${parts.hour}:${parts.minute}:${parts.second}`
}

// toShanghaiRFC3339 将日期选择器中的上海墙上时间转换为带 +08:00 的 RFC3339。
export const toShanghaiRFC3339 = (value) => {
  if (value === null || value === '' || typeof value === 'undefined') return undefined
  const text = String(value).trim()
  const localMatch = text.match(LOCAL_DATE_TIME_PATTERN)
  if (localMatch) {
    const [, year, month, day, hour, minute, second] = localMatch
    return `${year}-${month}-${day}T${hour}:${minute}:${second}${SHANGHAI_OFFSET}`
  }

  const formatted = formatOrderFoodDateTime(value)
  return formatted ? `${formatted.replace(' ', 'T')}${SHANGHAI_OFFSET}` : undefined
}

// getShanghaiPresetRange 生成上海自然日口径的今天、近 7 天或近 30 天筛选范围。
export const getShanghaiPresetRange = (rangeCode, now = new Date()) => {
  const days = { today: 1, last7Days: 7, last30Days: 30 }[rangeCode]
  if (!days) return []

  const current = formatOrderFoodDateTime(now)
  if (!current) return []
  const currentDate = current.slice(0, 10)
  const startAnchor = new Date(`${currentDate}T00:00:00Z`)
  startAnchor.setUTCDate(startAnchor.getUTCDate() - days + 1)
  return [`${startAnchor.toISOString().slice(0, 10)} 00:00:00`, current]
}
