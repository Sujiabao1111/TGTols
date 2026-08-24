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

        <el-form-item :label="TEXT.platform">
          <el-select
            v-model="searchInfo.platformCode"
            clearable
            :placeholder="TEXT.platformPlaceholder"
            class="w-[180px]"
          >
            <el-option label="HEDOC" value="HEDOC" />
            <el-option label="M7" value="M7" />
            <el-option label="M7PP" value="M7PP" />
          </el-select>
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
        <el-table-column align="left" :label="TEXT.startDate" min-width="180">
          <template #default="scope">
            <div class="whitespace-nowrap">{{ formatBeijingTime(scope.row.startDate || scope.row.createdAt) }}</div>
          </template>
        </el-table-column>
        <el-table-column align="left" :label="TEXT.userId" prop="userId" min-width="100" />
        <el-table-column align="left" :label="TEXT.username" prop="username" min-width="130" />
        <el-table-column align="left" :label="TEXT.platform" prop="platformCode" min-width="110" />
        <el-table-column align="left" :label="TEXT.provider" min-width="170">
          <template #default="scope">
            <div>{{ scope.row.providerName || '-' }}</div>
            <div class="text-xs text-gray-500">{{ scope.row.providerCode || scope.row.providerId || '-' }}</div>
          </template>
        </el-table-column>
        <el-table-column align="left" :label="TEXT.gameName" min-width="240">
          <template #default="scope">
            <div>{{ scope.row.gameNameCn || '-' }}</div>
            <div class="text-xs text-gray-500">{{ scope.row.gameNameEn || scope.row.gameCode || '-' }}</div>
          </template>
        </el-table-column>
        <el-table-column align="left" :label="TEXT.externalId" prop="externalId" min-width="210" show-overflow-tooltip />
        <el-table-column align="right" :label="TEXT.turnover" min-width="120">
          <template #default="scope">{{ formatAmount(scope.row.turnover) }}</template>
        </el-table-column>
        <el-table-column align="right" :label="TEXT.bet" min-width="120">
          <template #default="scope">{{ formatAmount(scope.row.bet) }}</template>
        </el-table-column>
        <el-table-column align="right" :label="TEXT.win" min-width="120">
          <template #default="scope">{{ formatAmount(scope.row.win) }}</template>
        </el-table-column>
        <el-table-column align="right" :label="TEXT.winLose" min-width="120">
          <template #default="scope">
            <span :class="Number(scope.row.winLose || 0) >= 0 ? 'text-green-600' : 'text-red-600'">
              {{ formatAmount(scope.row.winLose) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column align="left" :label="TEXT.endDate" min-width="180">
          <template #default="scope">
            <div class="whitespace-nowrap">{{ formatBeijingTime(scope.row.endDate) }}</div>
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
import { getUserGameRecordList } from '@/api/example/userGameRecord'

defineOptions({
  name: 'UserGameRecord'
})

const TEXT = {
  alert: '\u8f93\u5165\u73a9\u5bb6ID\u540e\uff0c\u53ef\u6309\u5355\u65e5\u6216\u65f6\u95f4\u6bb5\u67e5\u8be2\u8be5\u73a9\u5bb6\u7684\u6e38\u620f\u6295\u6ce8\u8bb0\u5f55\u3002',
  userId: '\u73a9\u5bb6ID',
  userIdPlaceholder: '\u8bf7\u8f93\u5165\u73a9\u5bb6ID',
  date: '\u6e38\u620f\u65e5\u671f',
  datePlaceholder: '\u9009\u62e9\u67d0\u5929',
  dateRange: '\u6e38\u620f\u65f6\u95f4\u6bb5',
  rangeSeparator: '\u81f3',
  startTime: '\u5f00\u59cb\u65f6\u95f4',
  endTime: '\u7ed3\u675f\u65f6\u95f4',
  platform: '\u6e38\u620f\u5e73\u53f0',
  platformPlaceholder: '\u9009\u62e9\u5e73\u53f0',
  username: '\u73a9\u5bb6\u8d26\u53f7',
  provider: '\u5382\u5546',
  gameName: '\u6e38\u620f\u540d\u79f0',
  externalId: '\u6295\u6ce8\u5355\u53f7',
  turnover: '\u6d41\u6c34',
  bet: '\u4e0b\u6ce8',
  win: '\u8d62',
  winLose: '\u8f93\u8d62',
  startDate: '\u5f00\u59cb\u65f6\u95f4',
  endDate: '\u7ed3\u675f\u65f6\u95f4',
  search: '\u67e5\u8be2',
  reset: '\u91cd\u7f6e',
  userIdRequired: '\u8bf7\u8f93\u5165\u73a9\u5bb6ID\u540e\u518d\u67e5\u8be2'
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
    dateRange: [],
    platformCode: ''
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

const resetTable = () => {
  tableData.value = []
  total.value = 0
}

const buildQuery = () => {
  const payload = {
    page: page.value,
    pageSize: pageSize.value,
    userId: searchInfo.value.userId,
    platformCode: searchInfo.value.platformCode
  }

  if (searchInfo.value.dateRange?.length === 2) {
    payload.dateRange = searchInfo.value.dateRange
  } else if (searchInfo.value.date) {
    payload.date = searchInfo.value.date
  }

  return payload
}

const getTableData = async() => {
  if (searchInfo.value.userId === undefined || searchInfo.value.userId === null || searchInfo.value.userId === '') {
    resetTable()
    return
  }

  const res = await getUserGameRecordList(buildQuery())
  if (res.code === 0) {
    tableData.value = res.data.list
    total.value = res.data.total
    page.value = res.data.page || page.value
    pageSize.value = res.data.pageSize || pageSize.value
  }
}

const onSubmit = async() => {
  if (searchInfo.value.userId === undefined || searchInfo.value.userId === null || searchInfo.value.userId === '') {
    ElMessage.warning(TEXT.userIdRequired)
    resetTable()
    return
  }

  page.value = 1
  await getTableData()
}

const onReset = () => {
  searchInfo.value = createDefaultSearchInfo()
  page.value = 1
  resetTable()
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
