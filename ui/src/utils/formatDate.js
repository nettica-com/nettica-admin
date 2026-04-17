import moment from 'moment'

export function formatDate(value) {
  if (!value) return ''
  return moment(String(value)).locale(navigator.language).format('L LT')
}
