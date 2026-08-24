<template>
  <div>
    <div class="gva-search-box">
      <el-form :inline="true" :model="searchInfo" @keyup.enter="onSubmit">
        <el-form-item :label="TEXT.userId">
          <el-input
            v-model.number="searchInfo.userId"
            clearable
            :placeholder="TEXT.userIdPlaceholder"
          />
        </el-form-item>

        <el-form-item :label="TEXT.username">
          <el-input
            v-model.trim="searchInfo.username"
            clearable
            :placeholder="TEXT.usernamePlaceholder"
          />
        </el-form-item>

        <el-form-item :label="TEXT.registerIp">
          <el-input
            v-model.trim="searchInfo.registerIp"
            clearable
            :placeholder="TEXT.registerIpPlaceholder"
          />
        </el-form-item>

        <el-form-item :label="TEXT.registerDomain">
          <el-select
            v-model="searchInfo.registerDomain"
            class="w-[210px]"
            clearable
            filterable
            allow-create
            default-first-option
            :placeholder="TEXT.registerDomainPlaceholder"
          >
            <el-option
              v-for="domain in DOMAIN_OPTIONS"
              :key="domain"
              :label="domain"
              :value="domain"
            />
          </el-select>
        </el-form-item>

        <el-form-item :label="TEXT.registerDate">
          <el-date-picker
            v-model="searchInfo.registerDate"
            type="date"
            value-format="YYYY-MM-DD"
            clearable
            :placeholder="TEXT.registerDatePlaceholder"
            @change="onRegisterDateChange"
          />
        </el-form-item>

        <el-form-item :label="TEXT.registerDateRange">
          <el-date-picker
            v-model="searchInfo.registerDateRange"
            class="w-[380px]"
            type="datetimerange"
            value-format="YYYY-MM-DD HH:mm:ss"
            :default-time="registerRangeDefaultTime"
            clearable
            :range-separator="TEXT.rangeSeparator"
            :start-placeholder="TEXT.startTime"
            :end-placeholder="TEXT.endTime"
            @change="onRegisterDateRangeChange"
          />
        </el-form-item>

        <el-form-item :label="TEXT.rechargeDate">
          <el-date-picker
            v-model="searchInfo.rechargeDate"
            type="date"
            value-format="YYYY-MM-DD"
            clearable
            :placeholder="TEXT.rechargeDatePlaceholder"
          />
        </el-form-item>

        <el-form-item :label="TEXT.rechargeDateRange">
          <el-date-picker
            v-model="searchInfo.rechargeDateRange"
            class="w-[380px]"
            type="datetimerange"
            value-format="YYYY-MM-DD HH:mm:ss"
            clearable
            :range-separator="TEXT.rangeSeparator"
            :start-placeholder="TEXT.startTime"
            :end-placeholder="TEXT.endTime"
          />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" icon="search" @click="onSubmit">{{ TEXT.search }}</el-button>
          <el-button icon="refresh" @click="onReset">{{ TEXT.reset }}</el-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="gva-table-box">
      <el-alert class="mb-4" type="info" :closable="false" :title="TEXT.alert" />

      <el-table :data="tableData" style="width: 100%" row-key="userId">
        <el-table-column align="left" :label="TEXT.userId" prop="userId" min-width="100" />
        <el-table-column align="left" :label="TEXT.username" prop="username" min-width="140" />
        <el-table-column align="left" :label="TEXT.balance" min-width="120">
          <template #default="scope">
            {{ formatAmount(scope.row.balance) }}
          </template>
        </el-table-column>
        <el-table-column align="left" :label="TEXT.totalTurnover" min-width="140">
          <template #default="scope">
            {{ formatAmount(scope.row.totalTurnover) }}
          </template>
        </el-table-column>
        <el-table-column align="left" :label="TEXT.totalWinlose" min-width="140">
          <template #default="scope">
            {{ formatAmount(scope.row.totalWinlose) }}
          </template>
        </el-table-column>
        <el-table-column align="left" :label="TEXT.rechargeCount" prop="rechargeCount" min-width="120" />
        <el-table-column align="left" :label="TEXT.rechargeAmount" min-width="140">
          <template #default="scope">
            {{ formatAmount(scope.row.rechargeAmount) }}
          </template>
        </el-table-column>
        <el-table-column align="left" :label="TEXT.withdrawAmount" min-width="140">
          <template #default="scope">
            {{ formatAmount(scope.row.withdrawAmount) }}
          </template>
        </el-table-column>
        <el-table-column align="left" :label="TEXT.totalGameCount" prop="totalGameCount" min-width="130" />
        <el-table-column align="left" :label="TEXT.slotGameCount" prop="slotGameCount" min-width="120" />
        <el-table-column align="left" :label="TEXT.casinoGameCount" prop="casinoGameCount" min-width="120" />
        <el-table-column align="left" :label="TEXT.sportbookGameCount" prop="sportbookGameCount" min-width="120" />
        <el-table-column align="left" :label="TEXT.otherGameCount" prop="otherGameCount" min-width="120" />
        <el-table-column align="left" :label="TEXT.createdAt" min-width="180">
          <template #default="scope">
            {{ formatBeijingTime(scope.row.createdAt) }}
          </template>
        </el-table-column>
        <el-table-column align="left" :label="TEXT.registerDomain" prop="registerDomain" min-width="160">
          <template #default="scope">
            {{ scope.row.registerDomain || '-' }}
          </template>
        </el-table-column>
        <el-table-column align="left" :label="TEXT.registerIp" prop="registerIp" min-width="160">
          <template #default="scope">
            {{ scope.row.registerIp || scope.row.lastLoginIp || '-' }}
          </template>
        </el-table-column>
        <el-table-column align="left" :label="TEXT.sameRegisterIpCount" prop="sameRegisterIpCount" min-width="130" />
        <el-table-column align="left" :label="TEXT.lastLoginAt" min-width="180">
          <template #default="scope">
            {{ formatBeijingTime(scope.row.lastLoginAt) }}
          </template>
        </el-table-column>
        <el-table-column align="left" :label="TEXT.lastLoginIp" min-width="180">
          <template #default="scope">
            <div>{{ scope.row.lastLoginIp || '-' }}</div>
            <div class="text-xs text-gray-500">{{ scope.row.ipCountryNote || TEXT.unknown }}</div>
          </template>
        </el-table-column>
      </el-table>

      <div class="gva-pagination">
        <el-pagination
          layout="total, sizes, prev, pager, next, jumper"
          :current-page="page"
          :page-size="pageSize"
          :page-sizes="[10, 30, 50, 100]"
          :total="total"
          @current-change="handleCurrentChange"
          @size-change="handleSizeChange"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { getUserDataList } from '@/api/example/userData'

defineOptions({
  name: 'UserData'
})

const TEXT = {
  registerDomain: '\u6ce8\u518c\u57df\u540d',
  registerDomainPlaceholder: '\u9009\u62e9\u6216\u8f93\u5165\u57df\u540d',
  registerIp: '注册IP',
  registerIpPlaceholder: '请输入注册IP',
  sameRegisterIpCount: '同IP注册数',
  rechargeDate: '\u5145\u503c\u65e5\u671f',
  rechargeDatePlaceholder: '\u9009\u62e9\u5145\u503c\u65e5\u671f',
  rechargeDateRange: '\u5145\u503c\u65f6\u95f4\u6bb5',
  alert: '默认展示当天注册的玩家数据。可按单天或某一时间段查询注册玩家，并支持分页浏览。',
  userId: '用户ID',
  userIdPlaceholder: '请输入用户ID',
  registerDate: '注册日期',
  registerDatePlaceholder: '选择某天',
  registerDateRange: '注册时间段',
  rangeSeparator: '至',
  startTime: '开始时间',
  endTime: '结束时间',
  username: '用户名',
  usernamePlaceholder: '请输入用户名',
  balance: '账户余额',
  totalTurnover: '总打码量',
  totalWinlose: '总输赢',
  rechargeCount: '充值次数',
  rechargeAmount: '充值总金额',
  withdrawAmount: '提现总金额',
  totalGameCount: '总游戏次数',
  slotGameCount: 'SLOT次数',
  casinoGameCount: '真人次数',
  sportbookGameCount: '体育次数',
  otherGameCount: '其他次数',
  createdAt: '注册时间',
  lastLoginAt: '最后登录时间',
  lastLoginIp: '最后登录IP',
  search: '查询',
  reset: '重置',
  unknown: '未知'
}

const DOMAIN_OPTIONS = [
  'ppnetpp.net',
  'ppnetpp.com',
  'ppbetpp.tech',
  'ppnet11.com',
  'ppnet22.com',
  'ppnet33.com',
  'ppnet44.com',
  'ppnet55.com'
]

const page = ref(1)
const total = ref(0)
const pageSize = ref(10)
const tableData = ref([])
const searchInfo = ref(createDefaultSearchInfo())
const registerRangeDefaultTime = [new Date(2000, 0, 1, 0, 0, 0), new Date(2000, 0, 1, 23, 59, 59)]

function createDefaultSearchInfo() {
  return {
    userId: undefined,
    username: '',
    registerIp: '',
    registerDomain: '',
    registerDate: currentDate(),
    registerDateRange: [],
    rechargeDate: undefined,
    rechargeDateRange: []
  }
}

function currentDate() {
  const formatter = new Intl.DateTimeFormat('en-CA', {
    timeZone: 'Asia/Shanghai',
    year: 'numeric',
    month: '2-digit',
    day: '2-digit'
  })
  return formatter.format(new Date())
}

const formatAmount = (value) => Number(value || 0).toFixed(2)

const onRegisterDateChange = (value) => {
  if (value) {
    searchInfo.value.registerDateRange = []
  }
}

const onRegisterDateRangeChange = (value) => {
  if (value?.length === 2) {
    searchInfo.value.registerDate = undefined
  }
}

const formatBeijingTime = (value) => {
  if (!value) {
    return '-'
  }

  if (typeof value === 'string') {
    const match = value.trim().match(/^(\d{4})-(\d{2})-(\d{2})[ T](\d{2}):(\d{2}):(\d{2})/)
    if (match) {
      return `${match[1]}-${match[2]}-${match[3]} ${match[4]}:${match[5]}:${match[6]}`
    }
  }

  const parsed = new Date(value)
  if (Number.isNaN(parsed.getTime())) {
    return String(value)
  }

  const formatter = new Intl.DateTimeFormat('zh-CN', {
    timeZone: 'Asia/Shanghai',
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false
  })

  const parts = formatter.formatToParts(parsed)
  const valueMap = {}
  for (const part of parts) {
    valueMap[part.type] = part.value
  }

  return `${valueMap.year}-${valueMap.month}-${valueMap.day} ${valueMap.hour}:${valueMap.minute}:${valueMap.second}`
}

const buildQuery = () => {
  const payload = {
    page: page.value,
    pageSize: pageSize.value,
    userId: searchInfo.value.userId,
    username: searchInfo.value.username,
    registerIp: searchInfo.value.registerIp,
    registerDomain: searchInfo.value.registerDomain
  }

  const hasUserId = searchInfo.value.userId !== undefined && searchInfo.value.userId !== null && searchInfo.value.userId !== ''
  const hasUsername = !!searchInfo.value.username
  const hasRegisterIp = !!searchInfo.value.registerIp
  const hasRechargeDateRange = searchInfo.value.rechargeDateRange?.length === 2
  const hasRechargeDate = !!searchInfo.value.rechargeDate
  const hasRechargeFilter = hasRechargeDateRange || hasRechargeDate

  if (hasRechargeDateRange) {
    payload.rechargeDateRange = searchInfo.value.rechargeDateRange
  } else if (hasRechargeDate) {
    payload.rechargeDate = searchInfo.value.rechargeDate
  }

  if (hasUserId || hasUsername || hasRegisterIp || hasRechargeFilter) {
    return payload
  }

  if (searchInfo.value.registerDateRange?.length === 2) {
    payload.registerDateRange = searchInfo.value.registerDateRange
  } else if (searchInfo.value.registerDate) {
    payload.registerDate = searchInfo.value.registerDate
  }

  return payload
}

const getTableData = async() => {
  const res = await getUserDataList(buildQuery())

  if (res.code === 0) {
    tableData.value = res.data.list
    total.value = res.data.total
    page.value = res.data.page || page.value
    pageSize.value = res.data.pageSize || pageSize.value
  }
}

const onSubmit = async() => {
  page.value = 1
  await getTableData()
}

const onReset = async() => {
  searchInfo.value = createDefaultSearchInfo()
  page.value = 1
  await getTableData()
}

const handleSizeChange = async(val) => {
  pageSize.value = val
  await getTableData()
}

const handleCurrentChange = async(val) => {
  page.value = val
  await getTableData()
}

getTableData()
</script>
