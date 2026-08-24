<template>
  <div class="activity-wager-config-page">
    <el-alert
      type="info"
      :closable="false"
      title="配置充值本金和赠送奖励部分的提现打码倍数，保存后前台提现校验会按新配置计算。"
    />

    <el-card class="mt-4" shadow="never">
      <template #header>
        <div class="flex items-center justify-between">
          <span class="font-semibold">打码配置</span>
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
        <el-form-item label="充值本金打码倍数" prop="depositWagerMultiplier">
          <el-input-number
            v-model="form.depositWagerMultiplier"
            :min="0.01"
            :step="0.5"
            :precision="2"
            controls-position="right"
            class="w-full"
          />
        </el-form-item>

        <el-form-item label="赠送奖励打码倍数" prop="rewardWagerMultiplier">
          <el-input-number
            v-model="form.rewardWagerMultiplier"
            :min="0.01"
            :step="0.5"
            :precision="2"
            controls-position="right"
            class="w-full"
          />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" :loading="saving" @click="submitForm">
            保存配置
          </el-button>
          <el-button @click="resetDefault">恢复默认 2x / 20x</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { ElMessage } from 'element-plus'
import { onMounted, reactive, ref } from 'vue'
import { getActivityWagerConfig, updateActivityWagerConfig } from '@/api/example/activityWagerConfig'

defineOptions({
  name: 'ActivityWagerConfig'
})

const formRef = ref()
const loading = ref(false)
const saving = ref(false)

const form = reactive({
  depositWagerMultiplier: 2,
  rewardWagerMultiplier: 20
})

const positiveRule = {
  required: true,
  type: 'number',
  min: 0.01,
  message: '请输入大于 0 的打码倍数',
  trigger: ['blur', 'change']
}

const rules = {
  depositWagerMultiplier: [positiveRule],
  rewardWagerMultiplier: [positiveRule]
}

const loadConfig = async() => {
  loading.value = true
  try {
    const res = await getActivityWagerConfig()
    if (res.code === 0 && res.data) {
      form.depositWagerMultiplier = Number(res.data.depositWagerMultiplier || 2)
      form.rewardWagerMultiplier = Number(res.data.rewardWagerMultiplier || 20)
    }
  } finally {
    loading.value = false
  }
}

const resetDefault = () => {
  form.depositWagerMultiplier = 2
  form.rewardWagerMultiplier = 20
  formRef.value?.clearValidate()
}

const submitForm = () => {
  formRef.value?.validate(async(valid) => {
    if (!valid) {
      return
    }

    saving.value = true
    try {
      const res = await updateActivityWagerConfig({
        depositWagerMultiplier: form.depositWagerMultiplier,
        rewardWagerMultiplier: form.rewardWagerMultiplier
      })
      if (res.code === 0) {
        ElMessage.success('保存成功')
        if (res.data) {
          form.depositWagerMultiplier = Number(res.data.depositWagerMultiplier || form.depositWagerMultiplier)
          form.rewardWagerMultiplier = Number(res.data.rewardWagerMultiplier || form.rewardWagerMultiplier)
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
