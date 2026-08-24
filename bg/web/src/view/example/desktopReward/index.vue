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

        <el-form-item :label="TEXT.grantContent">
          <el-input :model-value="TEXT.fixedGrantContent" disabled />
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
        <el-descriptions-item :label="TEXT.grantContent">
          {{ TEXT.fixedGrantContent }}
        </el-descriptions-item>
        <el-descriptions-item :label="TEXT.couponCode">
          {{ result.couponCode || '-' }}
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
import { grantDesktopReward } from '@/api/example/manualReward'

defineOptions({
  name: 'DesktopReward'
})

const TEXT = {
  alert: '\u8f93\u5165\u73a9\u5bb6 ID \u540e\uff0c\u7cfb\u7edf\u4f1a\u56fa\u5b9a\u53d1\u653e 1 \u5f20\u684c\u9762\u8865\u507f\u5238\u3002\u5982\u679c\u73a9\u5bb6\u5728\u67d0\u4e2a\u6e38\u620f\u4e2d\u8f93\u8d85\u8fc7\u81ea\u8eab\u5e26\u5165\u91d1\u989d\u7684 50% \u4ee5\u4e0a\uff0c\u5219\u53ef\u4ee5\u89e6\u53d1\u4fdd\u969c\u5238\uff0c\u5c06\u8865\u507f\u8f93\u6389\u90e8\u5206\u91d1\u989d\u7684 20%\u3002\u6bcf\u4e2a\u73a9\u5bb6 ID \u4ec5\u53ef\u53d1\u653e\u4e00\u6b21\u3002',
  userId: '\u73a9\u5bb6ID',
  userIdPlaceholder: '\u8bf7\u8f93\u5165\u73a9\u5bb6ID',
  grantContent: '\u53d1\u653e\u5185\u5bb9',
  fixedGrantContent: '\u8865\u507f\u5238 x1',
  submit: '\u53d1\u653e\u684c\u9762\u8865\u507f\u5238',
  reset: '\u91cd\u7f6e',
  resultTitle: '\u53d1\u653e\u7ed3\u679c',
  username: '\u73a9\u5bb6\u8d26\u53f7',
  couponCode: '\u8865\u507f\u5238\u7801',
  referenceId: '\u6d41\u6c34\u53c2\u8003ID'
}

const formRef = ref()
const loading = ref(false)
const result = ref(null)
const form = reactive({
  userId: undefined
})

const rules = {
  userId: [
    {
      required: true,
      message: '\u8bf7\u8f93\u5165\u73a9\u5bb6ID',
      trigger: ['blur', 'change']
    }
  ]
}

const resetForm = () => {
  form.userId = undefined
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
      const res = await grantDesktopReward({ userId: form.userId })
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
