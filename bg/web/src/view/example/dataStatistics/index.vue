<template>
  <div class="statistics-page" v-loading="loading">
    <section class="overview-grid">
      <article v-for="item in overviewCards" :key="item.key" class="metric-card">
        <div class="metric-icon" :class="item.tone"><el-icon><component :is="item.icon" /></el-icon></div>
        <div class="metric-copy">
          <span>{{ item.label }}</span>
          <strong>{{ item.money ? formatAmount(overview[item.key]) : formatInteger(overview[item.key]) }}</strong>
          <small>{{ item.hint }}</small>
        </div>
        <div class="spark" :class="item.tone"><i /><i /><i /><i /><i /></div>
      </article>
    </section>

    <section class="filter-band">
      <div>
        <h2>经营数据</h2>
        <p>{{ result.startDate }} 至 {{ result.endDate }}</p>
      </div>
      <div class="filter-actions">
        <el-radio-group v-model="quickRange" @change="applyQuickRange">
          <el-radio-button value="today">今日</el-radio-button>
          <el-radio-button value="yesterday">昨日</el-radio-button>
          <el-radio-button value="7days">近7天</el-radio-button>
          <el-radio-button value="30days">近30天</el-radio-button>
        </el-radio-group>
        <el-date-picker v-model="dateRange" type="daterange" value-format="YYYY-MM-DD" range-separator="至" start-placeholder="开始日期" end-placeholder="结束日期" :clearable="false" @change="onCustomRange" />
        <el-button type="primary" :icon="Refresh" @click="loadData">刷新</el-button>
      </div>
    </section>

    <section class="content-panel">
      <div class="panel-title"><h3>每日明细</h3><span>金额单位：U</span></div>
      <el-table :data="displayRows" stripe show-summary :summary-method="getSummaries">
        <el-table-column prop="date" label="日期" min-width="120" fixed />
        <el-table-column prop="registerUsers" label="注册人数" min-width="105" align="right" />
        <el-table-column prop="activeUsers" label="活跃人数" min-width="105" align="right" />
        <el-table-column prop="rechargeUsers" label="充值人数" min-width="105" align="right" />
        <el-table-column prop="rechargeAmount" label="充值金额" min-width="130" align="right"><template #default="{ row }">{{ formatAmount(row.rechargeAmount) }}</template></el-table-column>
        <el-table-column prop="withdrawUsers" label="提现人数" min-width="105" align="right" />
        <el-table-column prop="withdrawAmount" label="提现金额" min-width="130" align="right"><template #default="{ row }">{{ formatAmount(row.withdrawAmount) }}</template></el-table-column>
        <el-table-column prop="paymentGap" label="提现差" min-width="140" align="right"><template #default="{ row }"><span :class="row.paymentGap >= 0 ? 'positive' : 'negative'">{{ signedAmount(row.paymentGap) }}</span></template></el-table-column>
        <el-table-column prop="dailyWinLoss" label="当日总输赢" min-width="140" align="right"><template #default="{ row }"><span :class="row.dailyWinLoss >= 0 ? 'positive' : 'negative'">{{ signedAmount(row.dailyWinLoss) }}</span></template></el-table-column>
      </el-table>
    </section>
  </div>
</template>

<script setup>
import { computed, markRaw, onMounted, reactive, ref } from 'vue'
import { Coin, DataLine, Refresh, TrendCharts, User } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { getDataStatistics } from '@/api/example/dataStatistics'

defineOptions({ name: 'DataStatistics' })
const loading = ref(false)
const quickRange = ref('today')
const dateRange = ref([])
const result = reactive({ overview: {}, daily: [], startDate: '', endDate: '' })
const overview = computed(() => result.overview || {})
const overviewCards = [
  { key: 'totalRegisteredUsers', label: '总注册人数', hint: '平台累计注册用户', icon: markRaw(User), tone: 'blue' },
  { key: 'totalRechargeUsers', label: '总充值人数', hint: '至少成功充值一次', icon: markRaw(Coin), tone: 'green' },
  { key: 'totalRechargeAmount', label: '总充值金额', hint: '累计成功充值金额', icon: markRaw(TrendCharts), tone: 'amber', money: true },
  { key: 'totalWithdrawAmount', label: '总提现金额', hint: '累计成功提现金额', icon: markRaw(Coin), tone: 'purple', money: true },
  { key: 'totalPaymentGap', label: '总提充差', hint: '总充值金额 - 总提现金额', icon: markRaw(DataLine), tone: 'cyan', money: true },
  { key: 'todayWinLoss', label: '当日总输赢', hint: '今日平台游戏输赢', icon: markRaw(DataLine), tone: 'indigo', money: true },
  { key: 'platformWinLoss', label: '平台总输赢', hint: '平台累计游戏输赢', icon: markRaw(DataLine), tone: 'red', money: true }
]
const displayRows = computed(() => [...(result.daily || [])].reverse())
const pad = value => String(value).padStart(2, '0')
const formatDate = date => `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`
const setRange = (days, offset = 0) => { const end = new Date(); end.setDate(end.getDate() - offset); const start = new Date(end); start.setDate(end.getDate() - days + 1); dateRange.value = [formatDate(start), formatDate(end)] }
const applyQuickRange = value => { if (value === 'yesterday') setRange(1, 1); else setRange(value === '7days' ? 7 : value === '30days' ? 30 : 1); loadData() }
const onCustomRange = () => { quickRange.value = ''; loadData() }
const loadData = async () => {
  if (!dateRange.value?.length) return
  loading.value = true
  try { const res = await getDataStatistics({ startDate: dateRange.value[0], endDate: dateRange.value[1] }); Object.assign(result, res.data || {}) }
  catch (error) { ElMessage.error(error?.message || '获取数据统计失败') }
  finally { loading.value = false }
}
const formatInteger = value => Number(value || 0).toLocaleString('zh-CN')
const formatAmount = value => Number(value || 0).toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
const signedAmount = value => `${Number(value || 0) >= 0 ? '+' : ''}${formatAmount(value)}`
const getSummaries = ({ columns, data }) => columns.map((column, index) => { if (index === 0) return '合计'; const values = data.map(row => Number(row[column.property] || 0)); const total = values.reduce((sum, value) => sum + value, 0); return column.property?.includes('Amount') || column.property === 'paymentGap' || column.property === 'dailyWinLoss' ? formatAmount(total) : formatInteger(total) })
onMounted(() => { setRange(1); loadData() })
</script>

<style scoped lang="scss">
.statistics-page { padding: 16px; min-width: 0; }
.overview-grid { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 16px; margin-bottom: 16px; }
.metric-card { position: relative; display: flex; align-items: center; min-height: 126px; padding: 20px; overflow: hidden; background: #fff; border: 1px solid #edf0f5; border-radius: 8px; box-shadow: 0 2px 10px rgba(26, 39, 61, .05); }
.metric-icon { display: grid; place-items: center; width: 48px; height: 48px; flex: 0 0 48px; margin-right: 14px; border-radius: 8px; font-size: 23px; &.blue { color: #2563eb; background: #eaf2ff; } &.green { color: #059669; background: #e8f8f2; } &.amber { color: #d97706; background: #fff5df; } &.purple { color: #7c3aed; background: #f1eafe; } &.cyan { color: #0891b2; background: #e6f8fb; } &.indigo { color: #4f46e5; background: #ecebff; } &.red { color: #dc2626; background: #feebeb; } }
.metric-copy { min-width: 0; display: flex; flex-direction: column; z-index: 1; span { color: #687386; font-size: 14px; } strong { margin: 5px 0 4px; color: #182230; font-size: 25px; line-height: 1.2; font-weight: 700; white-space: nowrap; } small { color: #9aa3b2; } }
.spark { position: absolute; right: 12px; bottom: 19px; display: flex; align-items: flex-end; gap: 4px; height: 34px; opacity: .35; i { width: 5px; border-radius: 3px 3px 0 0; background: currentColor; } i:nth-child(1) { height: 8px } i:nth-child(2) { height: 15px } i:nth-child(3) { height: 12px } i:nth-child(4) { height: 25px } i:nth-child(5) { height: 31px } &.blue { color: #3b82f6 } &.green { color: #10b981 } &.amber { color: #f59e0b } &.purple { color: #8b5cf6 } &.cyan { color: #06b6d4 } &.indigo { color: #6366f1 } &.red { color: #ef4444 } }
.filter-band, .content-panel { background: #fff; border: 1px solid #edf0f5; border-radius: 8px; }
.filter-band { display: flex; justify-content: space-between; align-items: center; gap: 16px; padding: 17px 20px; margin-bottom: 16px; h2 { margin: 0; color: #1f2937; font-size: 18px; } p { margin: 4px 0 0; color: #8a94a6; font-size: 13px; } }
.filter-actions { display: flex; align-items: center; gap: 10px; flex-wrap: wrap; }
.content-panel { padding: 20px; margin-bottom: 16px; min-width: 0; }
.panel-title { display: flex; align-items: baseline; justify-content: space-between; margin-bottom: 14px; h3 { margin: 0; font-size: 16px; color: #1f2937; } span { color: #8a94a6; font-size: 12px; } }
.positive { color: #059669; font-weight: 600; } .negative { color: #dc2626; font-weight: 600; }
@media (max-width: 1180px) { .overview-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); } .filter-band { align-items: flex-start; flex-direction: column; } }
@media (max-width: 640px) { .statistics-page { padding: 10px; } .overview-grid { grid-template-columns: 1fr; gap: 10px; } .metric-card { min-height: 112px; } .filter-actions { width: 100%; } .filter-actions :deep(.el-date-editor) { width: 100%; } .content-panel { padding: 14px 10px; } }
</style>
