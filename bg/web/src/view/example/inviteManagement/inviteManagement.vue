<template>
  <div>
    <div class="gva-search-box">
      <el-form
        ref="elSearchFormRef"
        :inline="true"
        :model="searchInfo"
        class="demo-form-inline"
        @keyup.enter="onSubmit"
      >
        <el-form-item :label="TEXT.inviterUserId" prop="inviterUserId">
          <el-input
            v-model.number="searchInfo.inviterUserId"
            clearable
            :placeholder="TEXT.inviterUserIdPlaceholder"
          />
        </el-form-item>

        <el-form-item :label="TEXT.inviteCode" prop="inviteCode">
          <el-input
            v-model.trim="searchInfo.inviteCode"
            clearable
            :placeholder="TEXT.inviteCodePlaceholder"
          />
        </el-form-item>

        <el-form-item :label="TEXT.registerDomain" prop="registerDomain">
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

        <el-form-item :label="TEXT.registerDateRange" prop="statDateRange">
          <el-date-picker
            v-model="searchInfo.statDateRange"
            class="w-[300px]"
            type="daterange"
            value-format="YYYY-MM-DD"
            clearable
            :range-separator="TEXT.rangeSeparator"
            :start-placeholder="TEXT.startDate"
            :end-placeholder="TEXT.endDate"
          />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" icon="search" @click="onSubmit">{{ TEXT.search }}</el-button>
          <el-button icon="refresh" @click="onReset">{{ TEXT.reset }}</el-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="gva-table-box">
      <el-alert
        class="mb-4"
        type="info"
        :closable="false"
        :title="TEXT.alert"
      />

      <div class="invite-summary">
        <span>{{ TEXT.summaryPeriod }}</span>
        <span>{{ TEXT.summaryRegisterCount }}：{{ summaryData.registerCount }}</span>
        <span>{{ TEXT.summaryRechargeUserCount }}：{{ summaryData.rechargeUserCount }}</span>
        <span>{{ TEXT.summaryRechargeAmount }}：{{ formatAmount(summaryData.rechargeAmountU) }}</span>
      </div>

      <el-table
        style="width: 100%"
        tooltip-effect="dark"
        :data="tableData"
        row-key="userId"
      >
        <el-table-column align="left" :label="TEXT.inviterUserId" prop="inviterUserId" min-width="110" />
        <el-table-column align="left" :label="TEXT.userId" prop="userId" min-width="100" />
        <el-table-column align="left" :label="TEXT.username" prop="username" min-width="140" />
        <el-table-column align="right" :label="TEXT.balance" min-width="120">
          <template #default="scope">
            {{ formatAmount(scope.row.balance) }}
          </template>
        </el-table-column>
        <el-table-column align="left" :label="TEXT.rechargeCount" prop="rechargeCount" min-width="110" />
        <el-table-column align="right" :label="TEXT.rechargeAmount" min-width="140">
          <template #default="scope">
            {{ formatAmount(scope.row.rechargeAmount) }}
          </template>
        </el-table-column>
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
import { ElMessage } from 'element-plus'
import { ref } from 'vue'
import { getInviteStatsList } from '@/api/example/inviteManagement'

defineOptions({
  name: 'InviteManagement'
})

const TEXT = {
  registerDomain: '\u6ce8\u518c\u57df\u540d',
  registerDomainPlaceholder: '\u9009\u62e9\u6216\u8f93\u5165\u57df\u540d',
  inviterUserId: '邀请人ID',
  inviterUserIdPlaceholder: '请输入邀请人ID',
  inviteCode: '邀请码',
  inviteCodePlaceholder: '请输入邀请码',
  registerDateRange: '注册时间段',
  rangeSeparator: '至',
  startDate: '开始日期',
  endDate: '结束日期',
  search: '查询',
  reset: '重置',
  alert: '可按邀请人或注册域名查询用户明细，并按注册时间筛选。只选择域名时，会查询该域名注册的全部用户。',
  summaryPeriod: '时间段内',
  summaryRegisterCount: '总注册人数',
  summaryRechargeUserCount: '总充值人数',
  summaryRechargeAmount: '总充值金额(U)',
  userId: '用户ID',
  username: '用户名',
  balance: '账户余额',
  rechargeCount: '充值次数',
  rechargeAmount: '充值总金额',
  createdAt: '注册时间',
  lastLoginAt: '最后登录时间',
  lastLoginIp: '最后登录IP',
  inviterRequired: '请输入邀请人ID、邀请码或注册域名后再查询',
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

const elSearchFormRef = ref()
const page = ref(1)
const total = ref(0)
const pageSize = ref(10)
const tableData = ref([])
const summaryData = ref(createEmptySummary())
const searchInfo = ref(createDefaultSearchInfo())

function createDefaultSearchInfo() {
  return {
    inviterUserId: undefined,
    inviteCode: '',
    registerDomain: '',
    statDateRange: currentDateRange()
  }
}

function createEmptySummary() {
  return {
    registerCount: 0,
    rechargeUserCount: 0,
    rechargeAmountU: 0
  }
}

const formatAmount = (value) => Number(value || 0).toFixed(2)

function currentDate() {
  const formatter = new Intl.DateTimeFormat('en-CA', {
    timeZone: 'Asia/Shanghai',
    year: 'numeric',
    month: '2-digit',
    day: '2-digit'
  })
  return formatter.format(new Date())
}

function currentDateRange() {
  const today = currentDate()
  return [today, today]
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

const hasInviterFilter = () =>
  Number(searchInfo.value.inviterUserId || 0) > 0 ||
  !!String(searchInfo.value.inviteCode || '').trim() ||
  !!String(searchInfo.value.registerDomain || '').trim()

const resetTable = () => {
  tableData.value = []
  total.value = 0
  summaryData.value = createEmptySummary()
}

const buildQuery = () => {
  const payload = {
    page: page.value,
    pageSize: pageSize.value,
    inviteCode: searchInfo.value.inviteCode,
    registerDomain: searchInfo.value.registerDomain
  }

  if (Number(searchInfo.value.inviterUserId || 0) > 0) {
    payload.inviterUserId = searchInfo.value.inviterUserId
  }

  if (searchInfo.value.statDateRange?.length === 2) {
    payload.statDateRange = searchInfo.value.statDateRange
  }

  return payload
}

const getTableData = async() => {
  if (!hasInviterFilter()) {
    resetTable()
    return
  }

  const table = await getInviteStatsList(buildQuery())

  if (table.code === 0) {
    tableData.value = table.data.list
    total.value = table.data.total
    summaryData.value = table.data.summary || createEmptySummary()
    page.value = table.data.page || page.value
    pageSize.value = table.data.pageSize || pageSize.value
  }
}

const onReset = () => {
  searchInfo.value = createDefaultSearchInfo()
  page.value = 1
  resetTable()
}

const onSubmit = () => {
  elSearchFormRef.value?.validate(async(valid) => {
    if (!valid) return
    if (!hasInviterFilter()) {
      ElMessage.warning(TEXT.inviterRequired)
      resetTable()
      return
    }
    page.value = 1
    await getTableData()
  })
}

const handleSizeChange = async(val) => {
  pageSize.value = val
  await getTableData()
}

const handleCurrentChange = async(val) => {
  page.value = val
  await getTableData()
}
</script>

<style scoped>
.invite-summary {
  display: flex;
  flex-wrap: wrap;
  gap: 24px;
  align-items: center;
  min-height: 36px;
  padding: 0 12px;
  border: 1px solid #e4e7ed;
  border-bottom: 0;
  color: #1f2d3d;
  font-size: 14px;
  font-weight: 500;
}
</style>
