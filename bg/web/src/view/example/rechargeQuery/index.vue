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

        <el-form-item :label="TEXT.date">
          <el-date-picker
            v-model="searchInfo.date"
            type="date"
            value-format="YYYY-MM-DD"
            clearable
            :placeholder="TEXT.datePlaceholder"
          />
        </el-form-item>

        <el-form-item :label="TEXT.dateRange">
          <el-date-picker
            v-model="searchInfo.dateRange"
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

      <el-table :data="tableData" style="width: 100%" row-key="id">
        <el-table-column align="left" :label="TEXT.userId" prop="userId" min-width="100" />
        <el-table-column align="left" :label="TEXT.username" prop="username" min-width="140" />
        <el-table-column align="left" :label="TEXT.orderId" prop="orderId" min-width="210" show-overflow-tooltip />
        <el-table-column align="right" :label="TEXT.rechargeAmount" min-width="170">
          <template #default="scope">
            <div class="whitespace-nowrap">{{ formatLocalAmount(scope.row.localAmount, scope.row.currency) }}</div>
            <div class="text-xs text-gray-500">{{ formatAmount(scope.row.rechargeAmountU) }} U</div>
          </template>
        </el-table-column>
        <el-table-column align="right" :label="TEXT.totalRechargeAmount" min-width="150">
          <template #default="scope">
            {{ formatAmount(scope.row.totalRechargeAmount) }} U
          </template>
        </el-table-column>
        <el-table-column align="left" :label="TEXT.method" min-width="150">
          <template #default="scope">
            <div>{{ scope.row.dstCode || '-' }}</div>
            <div class="text-xs text-gray-500">{{ scope.row.type || '-' }}</div>
          </template>
        </el-table-column>
        <el-table-column align="left" :label="TEXT.paidAt" min-width="180">
          <template #default="scope">
            <div class="whitespace-nowrap">{{ formatBeijingTime(scope.row.paidAt) }}</div>
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
import { getRechargeQueryList } from '@/api/example/rechargeQuery'

defineOptions({
  name: 'RechargeQuery'
})

const TEXT = {
  alert: '默认展示当天所有充值成功的玩家。输入玩家ID后查询该玩家全部充值记录，也可按时间段查询所有成功充值记录。',
  userId: '玩家ID',
  userIdPlaceholder: '请输入玩家ID',
  date: '充值日期',
  datePlaceholder: '选择某天',
  dateRange: '充值时间段',
  rangeSeparator: '至',
  startTime: '开始时间',
  endTime: '结束时间',
  username: '玩家账号',
  orderId: '充值订单号',
  rechargeAmount: '充值金额',
  totalRechargeAmount: '玩家总充值',
  method: '充值方式',
  paidAt: '充值时间',
  search: '查询',
  reset: '重置'
}

const page = ref(1)
const total = ref(0)
const pageSize = ref(10)
const tableData = ref([])
const searchInfo = ref(createDefaultSearchInfo())

function createDefaultSearchInfo() {
  return {
    userId: undefined,
    date: currentDate(),
    dateRange: []
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

const hasUserId = () => searchInfo.value.userId !== undefined && searchInfo.value.userId !== null && searchInfo.value.userId !== ''

const formatAmount = (value) => Number(value || 0).toFixed(2)

const formatLocalAmount = (amount, currency) => {
  const unit = currency || ''
  return `${formatAmount(amount)} ${unit}`.trim()
}

const formatBeijingTime = (value) => {
  if (!value) {
    return '-'
  }

  const normalizedValue = typeof value === 'string' ? value.replace(' ', 'T') : value
  const parsed = new Date(normalizedValue)
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
    userId: searchInfo.value.userId
  }

  if (searchInfo.value.dateRange?.length === 2) {
    payload.dateRange = searchInfo.value.dateRange
  } else if (!hasUserId() && searchInfo.value.date) {
    payload.date = searchInfo.value.date
  }

  return payload
}

const getTableData = async() => {
  const res = await getRechargeQueryList(buildQuery())

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
