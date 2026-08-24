<template>
  <div class="space-y-4">
    <el-alert
      type="info"
      :closable="false"
      :title="TEXT.alert"
    />

    <el-card shadow="never" class="max-w-3xl">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="120px">
        <el-form-item :label="TEXT.linkLabel" prop="value">
          <el-input
            v-model.trim="form.value"
            :placeholder="TEXT.placeholder"
            clearable
          />
        </el-form-item>

        <el-form-item :label="TEXT.previewLabel">
          <el-link
            v-if="form.value"
            :href="form.value"
            type="primary"
            target="_blank"
          >
            {{ form.value }}
          </el-link>
          <span v-else class="text-gray-400">{{ TEXT.emptyPreview }}</span>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" :loading="loading" @click="saveConfig">
            {{ TEXT.save }}
          </el-button>
          <el-button @click="loadConfig">{{ TEXT.reload }}</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { ElMessage } from 'element-plus'
import { reactive, ref } from 'vue'
import { createSysParams, getSysParam, updateSysParams } from '@/api/sysParams'

defineOptions({
  name: 'WhatsappConfig'
})

const PARAM_KEY = 'WHATSAPP_URL'

const TEXT = {
  alert: '\u8bbe\u7f6e\u540e\uff0c\u524d\u53f0\u6240\u6709 WhatsApp \u5ba2\u670d\u5165\u53e3\u4f1a\u7edf\u4e00\u4f7f\u7528\u8fd9\u4e2a\u94fe\u63a5\u3002',
  linkLabel: 'WHATSAPP_URL',
  previewLabel: '\u9884\u89c8',
  placeholder: '\u4f8b\u5982: https://wa.me/628xxxxxxx',
  emptyPreview: '\u6682\u672a\u8bbe\u7f6e',
  save: '\u4fdd\u5b58',
  reload: '\u91cd\u65b0\u52a0\u8f7d'
}

const formRef = ref()
const loading = ref(false)
const recordId = ref(0)
const form = reactive({
  value: ''
})

const rules = {
  value: [
    {
      required: true,
      message: '\u8bf7\u8f93\u5165 WhatsApp \u94fe\u63a5',
      trigger: ['blur', 'change']
    },
    {
      validator: (_, value, callback) => {
        if (!value) {
          callback()
          return
        }
        try {
          const url = new URL(value)
          if (url.protocol !== 'https:') {
            callback(new Error('\u8bf7\u4f7f\u7528 https \u94fe\u63a5'))
            return
          }
          callback()
        } catch (error) {
          callback(new Error('\u8bf7\u8f93\u5165\u6709\u6548\u7684 URL'))
        }
      },
      trigger: ['blur', 'change']
    }
  ]
}

const loadConfig = async () => {
  loading.value = true
  try {
    const res = await getSysParam({ key: PARAM_KEY })
    if (res.code === 0 && res.data) {
      recordId.value = res.data.ID || 0
      form.value = res.data.value || ''
    } else {
      recordId.value = 0
      form.value = ''
    }
  } finally {
    loading.value = false
  }
}

const saveConfig = () => {
  formRef.value?.validate(async(valid) => {
    if (!valid) {
      return
    }

    loading.value = true
    try {
      const payload = {
        ID: recordId.value,
        name: 'WhatsApp Customer Service URL',
        key: PARAM_KEY,
        value: form.value,
        desc: 'Frontend customer support WhatsApp link'
      }

      const res = recordId.value
        ? await updateSysParams(payload)
        : await createSysParams(payload)

      if (res.code === 0) {
        ElMessage.success('\u4fdd\u5b58\u6210\u529f')
        await loadConfig()
      }
    } finally {
      loading.value = false
    }
  })
}

void loadConfig()
</script>
