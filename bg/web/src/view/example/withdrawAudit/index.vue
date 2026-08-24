<template>
  <div>
    <div class="gva-search-box">
      <el-form
        :inline="true"
        :model="searchInfo"
        @keyup.enter="onSubmit"
      >
        <el-form-item :label="TEXT.userId">
          <el-input
            v-model.number="searchInfo.userId"
            clearable
            :placeholder="TEXT.userIdPlaceholder"
          />
        </el-form-item>

        <el-form-item :label="TEXT.orderId">
          <el-input
            v-model.trim="searchInfo.orderId"
            clearable
            :placeholder="TEXT.orderIdPlaceholder"
          />
        </el-form-item>

        <el-form-item :label="TEXT.status">
          <el-select v-model="searchInfo.status" clearable :placeholder="TEXT.statusPlaceholder" class="w-[160px]">
            <el-option
              v-for="item in statusOptions"
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

      <el-table :data="tableData" style="width: 100%" row-key="orderId">
        <el-table-column align="left" :label="TEXT.createdAt" min-width="190">
          <template #default="scope">
            <div class="whitespace-nowrap">{{ formatBeijingTime(scope.row.createdAt) }}</div>
          </template>
        </el-table-column>
        <el-table-column align="left" :label="TEXT.orderId" prop="orderId" min-width="220" />
        <el-table-column align="left" :label="TEXT.userId" prop="userId" min-width="100" />
        <el-table-column align="left" :label="TEXT.username" prop="username" min-width="140" />
        <el-table-column align="left" :label="TEXT.amount" min-width="120">
          <template #default="scope">
            {{ formatAmount(scope.row.amount) }}
          </template>
        </el-table-column>
        <el-table-column align="left" :label="TEXT.cost" min-width="120">
          <template #default="scope">
            {{ formatAmount(scope.row.cost) }}
          </template>
        </el-table-column>
        <el-table-column align="left" :label="TEXT.method" min-width="120">
          <template #default="scope">
            {{ scope.row.dstCode }} / {{ scope.row.type }}
          </template>
        </el-table-column>
        <el-table-column align="left" :label="TEXT.account" min-width="180">
          <template #default="scope">
            <div>{{ scope.row.accountName }}</div>
            <div class="text-xs text-gray-500">{{ scope.row.account }}</div>
          </template>
        </el-table-column>
        <el-table-column align="left" :label="TEXT.status" min-width="120">
          <template #default="scope">
            <el-tag :type="statusTagType(scope.row.status)">
              {{ statusLabel(scope.row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column align="left" :label="TEXT.reviewInfo" min-width="180">
          <template #default="scope">
            <div>{{ scope.row.reviewer || '-' }}</div>
            <div class="whitespace-nowrap text-xs text-gray-500">{{ formatBeijingTime(scope.row.reviewedAt) }}</div>
          </template>
        </el-table-column>
        <el-table-column align="left" :label="TEXT.remark" prop="reviewRemark" min-width="180" show-overflow-tooltip />
        <el-table-column align="left" :label="TEXT.actions" fixed="right" min-width="180">
          <template #default="scope">
            <el-button
              v-if="scope.row.status === 0"
              link
              type="primary"
              @click="openReviewDialog('approve', scope.row)"
            >
              {{ TEXT.approve }}
            </el-button>
            <el-button
              v-if="scope.row.status === 0"
              link
              type="danger"
              @click="openReviewDialog('reject', scope.row)"
            >
              {{ TEXT.reject }}
            </el-button>
            <span v-else class="text-gray-400">-</span>
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

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="520px">
      <el-form label-width="90px">
        <el-form-item :label="TEXT.orderId">
          <span>{{ currentRow?.orderId || '-' }}</span>
        </el-form-item>
        <el-form-item :label="TEXT.userId">
          <span>{{ currentRow?.userId || '-' }}</span>
        </el-form-item>
        <el-form-item :label="TEXT.amount">
          <span>{{ formatAmount(currentRow?.amount) }}</span>
        </el-form-item>
        <el-form-item :label="TEXT.remark">
          <el-input
            v-model.trim="reviewRemark"
            type="textarea"
            :rows="4"
            :placeholder="TEXT.remarkPlaceholder"
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">{{ TEXT.cancel }}</el-button>
        <el-button :type="dialogAction === 'approve' ? 'primary' : 'danger'" :loading="submitLoading" @click="submitReview">
          {{ dialogAction === 'approve' ? TEXT.confirmApprove : TEXT.confirmReject }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ElMessage } from 'element-plus'
import { computed, ref } from 'vue'
import { approveWithdrawOrder, getWithdrawAuditList, rejectWithdrawOrder } from '@/api/example/withdrawAudit'

defineOptions({
  name: 'WithdrawAudit'
})

const TEXT = {
  alert: '玩家提现申请会先进入待审核，审核通过后才会记为提现成功，驳回会自动退回余额。',
  userId: '玩家ID',
  userIdPlaceholder: '请输入玩家ID',
  username: '玩家账号',
  orderId: '提现单号',
  orderIdPlaceholder: '请输入提现单号',
  status: '状态',
  statusPlaceholder: '请选择状态',
  amount: '提现金额',
  cost: '扣减U额',
  method: '提现方式',
  account: '收款账户',
  createdAt: '申请时间',
  reviewInfo: '审核信息',
  remark: '审核备注',
  actions: '操作',
  search: '查询',
  reset: '重置',
  approve: '通过',
  reject: '驳回',
  cancel: '取消',
  confirmApprove: '确认通过',
  confirmReject: '确认驳回',
  approveTitle: '审核通过提现',
  rejectTitle: '驳回提现申请',
  remarkPlaceholder: '可填写审核备注，供后台留档'
}

const statusOptions = [
  { label: '待审核', value: 0 },
  { label: '处理中', value: 3 },
  { label: '已通过', value: 1 },
  { label: '已驳回', value: 2 }
]

const page = ref(1)
const total = ref(0)
const pageSize = ref(10)
const tableData = ref([])
const searchInfo = ref({
  userId: undefined,
  orderId: '',
  status: undefined
})

const dialogVisible = ref(false)
const dialogAction = ref('approve')
const currentRow = ref(null)
const reviewRemark = ref('')
const submitLoading = ref(false)

const dialogTitle = computed(() => (dialogAction.value === 'approve' ? TEXT.approveTitle : TEXT.rejectTitle))

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

const statusLabel = (status) => {
  const matched = statusOptions.find((item) => item.value === status)
  return matched ? matched.label : `状态${status}`
}

const statusTagType = (status) => {
  if (status === 1) return 'success'
  if (status === 2) return 'danger'
  if (status === 3) return 'primary'
  return 'warning'
}

const getTableData = async() => {
  const res = await getWithdrawAuditList({
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
  page.value = 1
  await getTableData()
}

const onReset = async() => {
  searchInfo.value = {
    userId: undefined,
    orderId: '',
    status: undefined
  }
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

const openReviewDialog = (action, row) => {
  dialogAction.value = action
  currentRow.value = row
  reviewRemark.value = ''
  dialogVisible.value = true
}

const submitReview = async() => {
  if (!currentRow.value?.orderId) {
    return
  }

  submitLoading.value = true
  try {
    const api = dialogAction.value === 'approve' ? approveWithdrawOrder : rejectWithdrawOrder
    const res = await api({
      orderId: currentRow.value.orderId,
      remark: reviewRemark.value
    })
    if (res.code === 0) {
      ElMessage.success(dialogAction.value === 'approve' ? '审核通过，已提交三方代付' : '驳回成功')
      dialogVisible.value = false
      await getTableData()
    }
  } finally {
    submitLoading.value = false
  }
}

getTableData()
</script>
