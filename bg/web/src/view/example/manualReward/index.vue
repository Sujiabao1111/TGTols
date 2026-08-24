<template>
  <div class="space-y-4">
    <el-alert
      type="info"
      :closable="false"
      :title="TEXT.alert"
    />

    <div class="gva-search-box">
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="100px"
        class="max-w-xl"
      >
        <el-form-item :label="TEXT.userId" prop="userId">
          <el-input-number
            v-model="form.userId"
            :min="1"
            :step="1"
            class="w-full"
            :placeholder="TEXT.userIdPlaceholder"
            controls-position="right"
          />
        </el-form-item>

        <el-form-item :label="TEXT.rewardValue" prop="rewardAmountU">
          <el-input-number
            v-model="form.rewardAmountU"
            :min="0.01"
            :step="1"
            :precision="2"
            class="w-full"
            :placeholder="TEXT.rewardPlaceholder"
            controls-position="right"
          />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" :loading="loading" @click="submitForm">
            {{ TEXT.submit }}
          </el-button>
          <el-button @click="resetForm">{{ TEXT.reset }}</el-button>
        </el-form-item>
      </el-form>
    </div>

    <el-card v-if="result" shadow="never">
      <template #header>
        <span>{{ TEXT.resultTitle }}</span>
      </template>

      <el-descriptions :column="2" border>
        <el-descriptions-item :label="TEXT.userId">
          {{ result.userId }}
        </el-descriptions-item>
        <el-descriptions-item :label="TEXT.username">
          {{ result.username }}
        </el-descriptions-item>
        <el-descriptions-item :label="TEXT.rewardValue">
          {{ formatNumber(result.rewardAmountU) }}U
        </el-descriptions-item>
        <el-descriptions-item :label="TEXT.currency">
          {{ result.currency }}
        </el-descriptions-item>
        <el-descriptions-item :label="TEXT.localAmount">
          {{ formatNumber(result.localAmount) }}
        </el-descriptions-item>
        <el-descriptions-item :label="TEXT.exchangeRate">
          {{ formatNumber(result.exchangeRate) }}
        </el-descriptions-item>
        <el-descriptions-item :label="TEXT.wagerMultiplier">
          {{ formatNumber(result.wagerMultiplier) }}x
        </el-descriptions-item>
        <el-descriptions-item :label="TEXT.wagerRequired">
          {{ formatNumber(result.wagerRequired) }}U
        </el-descriptions-item>
        <el-descriptions-item :label="TEXT.beforeBalance">
          {{ formatNumber(result.beforeBalance) }}U
        </el-descriptions-item>
        <el-descriptions-item :label="TEXT.afterBalance">
          {{ formatNumber(result.afterBalance) }}U
        </el-descriptions-item>
        <el-descriptions-item :label="TEXT.referenceId" :span="2">
          {{ result.referenceId }}
        </el-descriptions-item>
      </el-descriptions>
    </el-card>
  </div>
</template>

<script setup>
import { ElMessage } from 'element-plus'
import { reactive, ref } from 'vue'
import { grantReward } from '@/api/example/manualReward'

defineOptions({
  name: 'ManualReward'
})

const TEXT = {
  alert: '\u8f93\u5165\u73a9\u5bb6 ID \u548c U \u5355\u4f4d\u5956\u52b1\u6570\u503c\uff0c\u7cfb\u7edf\u4f1a\u81ea\u52a8\u5165\u8d26\uff0c\u5e76\u6309\u3010\u914d\u7f6e\u9879-\u6253\u7801\u914d\u7f6e\u3011\u4e2d\u7684\u8d60\u9001\u5956\u52b1\u6253\u7801\u500d\u6570\u8bb0\u5f55\u6253\u7801\u91cf\uff0c\u5b8c\u6210\u540e\u624d\u53ef\u63d0\u73b0\u3002',
  userId: '\u73a9\u5bb6ID',
  userIdPlaceholder: '\u8bf7\u8f93\u5165\u73a9\u5bb6ID',
  rewardValue: '\u5956\u52b1\u6570\u503c(U)',
  rewardPlaceholder: '\u8bf7\u8f93\u5165\u5956\u52b1U\u503c',
  submit: '\u53d1\u653e\u5956\u52b1',
  reset: '\u91cd\u7f6e',
  resultTitle: '\u53d1\u653e\u7ed3\u679c',
  username: '\u73a9\u5bb6\u8d26\u53f7',
  currency: '\u5bf9\u5e94\u5e01\u79cd',
  localAmount: '\u6362\u7b97\u91d1\u989d',
  exchangeRate: '\u5f53\u524d\u6c47\u7387',
  wagerMultiplier: '\u6253\u7801\u500d\u6570',
  wagerRequired: '\u6240\u9700\u6253\u7801\u91cf',
  beforeBalance: '\u53d1\u653e\u524d\u4f59\u989d',
  afterBalance: '\u53d1\u653e\u540e\u4f59\u989d',
  referenceId: '\u6d41\u6c34\u53c2\u8003ID'
}

const formRef = ref()
const loading = ref(false)
const result = ref(null)
const form = reactive({
  userId: undefined,
  rewardAmountU: undefined
})

const rules = {
  userId: [
    {
      required: true,
      message: '\u8bf7\u8f93\u5165\u73a9\u5bb6ID',
      trigger: ['blur', 'change']
    }
  ],
  rewardAmountU: [
    {
      required: true,
      message: '\u8bf7\u8f93\u5165\u5956\u52b1\u6570\u503c',
      trigger: ['blur', 'change']
    }
  ]
}

const formatNumber = (value) => Number(value || 0).toFixed(2)

const resetForm = () => {
  form.userId = undefined
  form.rewardAmountU = undefined
  result.value = null
  formRef.value?.clearValidate()
}

const submitForm = () => {
  formRef.value?.validate(async(valid) => {
    if (!valid) {
      return
    }

    loading.value = true
    try {
      const res = await grantReward({
        userId: form.userId,
        rewardAmountU: form.rewardAmountU
      })
      if (res.code === 0) {
        result.value = res.data
        ElMessage.success('\u53d1\u653e\u6210\u529f')
      }
    } finally {
      loading.value = false
    }
  })
}
</script>
