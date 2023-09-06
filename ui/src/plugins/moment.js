import Vue from 'vue'
import moment from 'moment';
import VueMoment from 'vue-moment'

moment.locale('en');

Vue.use(VueMoment, {
  moment
});
// $moment() accessible in project

Vue.filter('formatDate', function (value) {
  if (!value) return '';
//  return moment(String(value)).format('MM/DD/YYYY hh:mm A');
  return moment(String(value)).locale(navigator.language).format("L LT");
});
