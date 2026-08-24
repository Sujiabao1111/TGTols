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
        <el-form-item prop="statDateRange">
          <template #label>
            <span>
              统计日期(注册日期)
              <el-tooltip content="搜索范围为开始日期(包含)至结束日期(包含)">
                <el-icon><QuestionFilled /></el-icon>
              </el-tooltip>
            </span>
          </template>
          <el-date-picker
            v-model="searchInfo.statDateRange"
            class="w-[380px]"
            type="daterange"
            range-separator="至"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            value-format="YYYY-MM-DD"
          />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" :icon="Search" @click="onSubmit">查询</el-button>
          <el-button :icon="Refresh" @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="gva-table-box">
      <el-table
        style="width: 100%"
        tooltip-effect="dark"
        :data="tableData"
        row-key="id"
      >
        <el-table-column align="left" label="ID" prop="id" width="90" />
        <el-table-column align="left" label="统计日期(注册日期)" prop="statDate" width="160">
          <template #default="scope">{{ formatDate(scope.row.statDate).slice(0, 10) }}</template>
        </el-table-column>
        <el-table-column align="left" label="当日新增注册人数" prop="newUsers" width="130" />
        <el-table-column align="left" label="当日充值人数" prop="depositUsers" width="120" />
        <el-table-column align="left" label="当日充值总金额" prop="depositAmount" width="140">
          <template #default="scope">{{ formatAmount(scope.row.depositAmount) }}</template>
        </el-table-column>
        <el-table-column align="left" label="当日有输赢玩家数" prop="gameWinloseUsers" width="140" />
        <el-table-column align="left" label="次日留存数" prop="retention1" width="110" />
        <el-table-column align="left" label="3日留存数" prop="retention3" width="110" />
        <el-table-column align="left" label="7日留存数" prop="retention7" width="110" />
        <el-table-column align="left" label="月留存数" prop="retention30" width="110" />
        <el-table-column align="left" label="操作" fixed="right" :min-width="appStore.operateMinWith">
          <template #default="scope">
            <el-button type="primary" link class="table-button" @click="getDetails(scope.row)">
              <el-icon style="margin-right: 5px"><InfoFilled /></el-icon>查看
            </el-button>
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

    <el-drawer
      v-model="detailShow"
      destroy-on-close
      :size="appStore.drawerSize"
      :show-close="true"
      :before-close="closeDetailShow"
      title="查看"
    >
      <el-descriptions :column="1" border>
        <el-descriptions-item label="ID">{{ detailFrom.id }}</el-descriptions-item>
        <el-descriptions-item label="统计日期(注册日期)">
          {{ formatDate(detailFrom.statDate).slice(0, 10) }}
        </el-descriptions-item>
        <el-descriptions-item label="当日新增注册人数">{{ detailFrom.newUsers }}</el-descriptions-item>
        <el-descriptions-item label="当日充值人数">{{ detailFrom.depositUsers }}</el-descriptions-item>
        <el-descriptions-item label="当日充值总金额">
          {{ formatAmount(detailFrom.depositAmount) }}
        </el-descriptions-item>
        <el-descriptions-item label="当日有输赢玩家数">{{ detailFrom.gameWinloseUsers }}</el-descriptions-item>
        <el-descriptions-item label="次日留存数">{{ detailFrom.retention1 }}</el-descriptions-item>
        <el-descriptions-item label="3日留存数">{{ detailFrom.retention3 }}</el-descriptions-item>
        <el-descriptions-item label="7日留存数">{{ detailFrom.retention7 }}</el-descriptions-item>
        <el-descriptions-item label="月留存数">{{ detailFrom.retention30 }}</el-descriptions-item>
      </el-descriptions>
    </el-drawer>
  </div>
</template>

<script setup>
import { InfoFilled, QuestionFilled, Refresh, Search } from '@element-plus/icons-vue'
import { ref } from 'vue'
import { useAppStore } from '@/pinia'
import { formatDate } from '@/utils/format'
import { findStatsRetention, getStatsRetentionList } from '@/api/example/statsRetention'

defineOptions({
  name: 'StatsRetention'
})

const appStore = useAppStore()
const elSearchFormRef = ref()

const page = ref(1)
const total = ref(0)
const pageSize = ref(10)
const tableData = ref([])
const searchInfo = ref({})

const detailShow = ref(false)
const detailFrom = ref({})

const formatAmount = (value) => {
  const amount = Number(value ?? 0)
  return Number.isFinite(amount) ? amount.toFixed(2) : '0.00'
}

const getTableData = async () => {
  const table = await getStatsRetentionList({
    page: page.value,
    pageSize: pageSize.value,
    ...searchInfo.value
  })

  if (table.code === 0) {
    tableData.value = table.data.list
    total.value = table.data.total
    page.value = table.data.page
    pageSize.value = table.data.pageSize
  }
}

const onReset = () => {
  searchInfo.value = {}
  page.value = 1
  getTableData()
}

const onSubmit = () => {
  page.value = 1
  getTableData()
}

const handleSizeChange = (val) => {
  pageSize.value = val
  getTableData()
}

const handleCurrentChange = (val) => {
  page.value = val
  getTableData()
}

const getDetails = async (row) => {
  const res = await findStatsRetention({ id: row.id })
  if (res.code === 0) {
    detailFrom.value = res.data
    detailShow.value = true
  }
}

const closeDetailShow = () => {
  detailShow.value = false
  detailFrom.value = {}
}

getTableData()
</script>
