<template>
  <div class="withdraw-fee-config-page">
    <el-alert
      type="info"
      :closable="false"
      title="配置玩家提现提交三方前的平台手续费。开启后，三方实付金额 = 提现金额 × (1 - 手续费率)；后台驳回仍按原扣款全额退回玩家。"
    />

    <el-card class="mt-4" shadow="never">
      <template #header>
        <div class="flex items-center justify-between">
          <span class="font-semibold">提现手续费配置</span>
          <el-button :loading="loading" @click="loadConfig">刷新</el-button>
        </div>
      </template>

      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="180px"
        class="max-w-xl"
      >
        <el-form-item label="启用手续费" prop="enabled">
          <el-switch
            v-model="form.enabled"
            active-text="开启"
            inactive-text="关闭"
          />
        </el-form-item>

        <el-form-item label="手续费率" prop="feeRate">
          <el-input-number
            v-model="form.feeRate"
            :min="0"
            :max="1"
            :step="0.001"
            :precision="6"
            controls-position="right"
            class="w-full"
          />
          <div class="mt-1 text-xs text-gray-500">
            当前约为 {{ feePercent }}%，默认 0.003 = 0.3%。
          </div>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" :loading="saving" @click="submitForm">
            保存配置
          </el-button>
          <el-button @click="resetDefault">恢复默认开启 0.3%</el-button>
          <el-button @click="disableFee">关闭手续费</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { ElMessage } from 'element-plus'
import { computed, onMounted, reactive, ref } from 'vue'
import { getWithdrawFeeConfig, updateWithdrawFeeConfig } from '@/api/example/withdrawFeeConfig'

defineOptions({
  name: 'WithdrawFeeConfig'
})

const formRef = ref()
const loading = ref(false)
const saving = ref(false)

const form = reactive({
  enabled: true,
  feeRate: 0.003
})

const feePercent = computed(() => (Number(form.feeRate || 0) * 100).toFixed(4))

const rules = {
  feeRate: [
    {
      required: true,
      type: 'number',
      min: 0,
      max: 1,
      message: '请输入 0 到 1 之间的手续费率',
      trigger: ['blur', 'change']
    }
  ]
}

const loadConfig = async() => {
  loading.value = true
  try {
    const res = await getWithdrawFeeConfig()
    if (res.code === 0 && res.data) {
      form.enabled = Boolean(res.data.enabled)
      form.feeRate = Number(res.data.feeRate ?? 0.003)
    }
  } finally {
    loading.value = false
  }
}

const resetDefault = () => {
  form.enabled = true
  form.feeRate = 0.003
  formRef.value?.clearValidate()
}

const disableFee = () => {
  form.enabled = false
  formRef.value?.clearValidate()
}

const submitForm = () => {
  formRef.value?.validate(async(valid) => {
    if (!valid) {
      return
    }

    saving.value = true
    try {
      const res = await updateWithdrawFeeConfig({
        enabled: form.enabled,
        feeRate: Number(form.feeRate || 0)
      })
      if (res.code === 0) {
        ElMessage.success('保存成功')
        if (res.data) {
          form.enabled = Boolean(res.data.enabled)
          form.feeRate = Number(res.data.feeRate ?? form.feeRate)
        }
      }
    } finally {
      saving.value = false
    }
  })
}

onMounted(() => {
  void loadConfig()
})
</script>
