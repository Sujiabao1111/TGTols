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

        <el-form-item :label="TEXT.errorType">
          <el-select
            v-model="searchInfo.errorType"
            clearable
            :placeholder="TEXT.errorTypePlaceholder"
            class="w-[220px]"
          >
            <el-option
              v-for="item in errorTypeOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
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
        <el-table-column align="left" :label="TEXT.createdAt" min-width="180">
          <template #default="scope">
            <div class="whitespace-nowrap">{{ formatBeijingTime(scope.row.createdAt) }}</div>
          </template>
        </el-table-column>
        <el-table-column align="left" :label="TEXT.userId" prop="userId" min-width="100" />
        <el-table-column align="left" :label="TEXT.username" prop="username" min-width="140" />
        <el-table-column align="left" :label="TEXT.gameInfo" min-width="220">
          <template #default="scope">
            <div>{{ scope.row.gameName || '-' }}</div>
            <div class="text-xs text-gray-500">{{ scope.row.gameCode || '-' }}</div>
          </template>
        </el-table-column>
        <el-table-column align="left" :label="TEXT.providerInfo" min-width="180">
          <template #default="scope">
            <div>{{ scope.row.providerName || '-' }}</div>
            <div class="text-xs text-gray-500">{{ scope.row.providerCode || '-' }}</div>
          </template>
        </el-table-column>
        <el-table-column align="left" :label="TEXT.errorType" min-width="160">
          <template #default="scope">
            <el-tag :type="errorTypeTagType(scope.row.errorType)">
              {{ errorTypeLabel(scope.row.errorType) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column align="left" :label="TEXT.errorMessage" prop="errorMessage" min-width="260" show-overflow-tooltip />
        <el-table-column align="left" :label="TEXT.scene" min-width="150">
          <template #default="scope">
            <div>{{ scope.row.isLobby ? TEXT.lobby : TEXT.directGame }}</div>
            <div class="text-xs text-gray-500">
              {{ scope.row.isMobile ? TEXT.mobile : TEXT.desktop }} / {{ scope.row.language || '-' }}
            </div>
          </template>
        </el-table-column>
        <el-table-column align="left" :label="TEXT.client" min-width="180">
          <template #default="scope">
            <div>{{ scope.row.clientIp || '-' }}</div>
            <div class="text-xs text-gray-500" :title="scope.row.userAgent || ''">{{ scope.row.userAgent || '-' }}</div>
          </template>
        </el-table-column>
        <el-table-column align="left" :label="TEXT.pageUrl" prop="pageUrl" min-width="240" show-overflow-tooltip />
        <el-table-column align="left" :label="TEXT.gameUrlHost" prop="gameUrlHost" min-width="180" show-overflow-tooltip />
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
import { getGameLaunchErrorList } from '@/api/example/gameLaunchError'

defineOptions({
  name: 'GameLaunchError'
})

const TEXT = {
  alert: '这里只记录玩家点击进入三方游戏后未正常进入、卡住或黑屏时上报的异常信息，正常进入不会记录。',
  userId: '玩家ID',
  userIdPlaceholder: '请输入玩家ID',
  username: '玩家账号',
  gameInfo: '游戏信息',
  providerInfo: '厂商信息',
  errorType: '报错类型',
  errorTypePlaceholder: '请选择报错类型',
  errorMessage: '报错信息',
  scene: '进入场景',
  client: '客户端信息',
  pageUrl: '页面地址',
  gameUrlHost: '三方域名',
  createdAt: '报错时间',
  search: '查询',
  reset: '重置',
  userIdRequired: '请输入玩家ID后再查询',
  lobby: '大厅进入',
  directGame: '直进游戏',
  mobile: '移动端',
  desktop: '桌面端'
}

const errorTypeOptions = [
  { label: '请求三方失败', value: 'launch_request_failed' },
  { label: 'iframe 加载超时', value: 'iframe_load_timeout' }
]

const page = ref(1)
const total = ref(0)
const pageSize = ref(10)
const tableData = ref([])
const searchInfo = ref({
  userId: undefined,
  errorType: ''
})

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

const errorTypeLabel = (value) => {
  const matched = errorTypeOptions.find((item) => item.value === value)
  return matched ? matched.label : value || '-'
}

const errorTypeTagType = (value) => {
  if (value === 'iframe_load_timeout') {
    return 'warning'
  }
  if (value === 'launch_request_failed') {
    return 'danger'
  }
  return 'info'
}

const resetTable = () => {
  tableData.value = []
  total.value = 0
}

const getTableData = async() => {
  if (searchInfo.value.userId === undefined || searchInfo.value.userId === null || searchInfo.value.userId === '') {
    resetTable()
    return
  }

  const res = await getGameLaunchErrorList({
    page: page.value,
    pageSize: pageSize.value,
    ...searchInfo.value
  })

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
  searchInfo.value = {
    userId: undefined,
    errorType: ''
  }
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
