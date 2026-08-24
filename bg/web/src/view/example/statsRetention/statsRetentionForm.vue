<template>
  <div>
    <div class="gva-form-box">
      <el-form ref="elFormRef" :model="formData" label-position="right" :rules="rule" label-width="120px">
        <el-form-item label="ID" prop="id">
          <el-input v-model.number="formData.id" :clearable="true" placeholder="请输入 ID" />
        </el-form-item>
        <el-form-item label="统计日期" prop="statDate">
          <el-date-picker v-model="formData.statDate" type="date" style="width: 100%" placeholder="选择日期" :clearable="true" />
        </el-form-item>
        <el-form-item label="当日新增注册人数" prop="newUsers">
          <el-input v-model.number="formData.newUsers" :clearable="true" placeholder="请输入当日新增注册人数" />
        </el-form-item>
        <el-form-item label="当日充值人数" prop="depositUsers">
          <el-input v-model.number="formData.depositUsers" :clearable="true" placeholder="请输入当日充值人数" />
        </el-form-item>
        <el-form-item label="当日充值总金额" prop="depositAmount">
          <el-input v-model.number="formData.depositAmount" :clearable="true" placeholder="请输入当日充值总金额" />
        </el-form-item>
        <el-form-item label="当日有输赢玩家数" prop="gameWinloseUsers">
          <el-input v-model.number="formData.gameWinloseUsers" :clearable="true" placeholder="请输入当日有输赢玩家数" />
        </el-form-item>
        <el-form-item label="次日留存数" prop="retention1">
          <el-input v-model.number="formData.retention1" :clearable="true" placeholder="请输入次日留存数" />
        </el-form-item>
        <el-form-item label="3日留存数" prop="retention3">
          <el-input v-model.number="formData.retention3" :clearable="true" placeholder="请输入3日留存数" />
        </el-form-item>
        <el-form-item label="7日留存数" prop="retention7">
          <el-input v-model.number="formData.retention7" :clearable="true" placeholder="请输入7日留存数" />
        </el-form-item>
        <el-form-item label="月留存数" prop="retention30">
          <el-input v-model.number="formData.retention30" :clearable="true" placeholder="请输入月留存数" />
        </el-form-item>
        <el-form-item>
          <el-button :loading="btnLoading" type="primary" @click="save">保存</el-button>
          <el-button type="primary" @click="back">返回</el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script setup>
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ref, reactive } from 'vue'
import {
  createStatsRetention,
  updateStatsRetention,
  findStatsRetention
} from '@/api/example/statsRetention'

defineOptions({
  name: 'StatsRetentionForm'
})

const route = useRoute()
const router = useRouter()
const btnLoading = ref(false)
const type = ref('')
const elFormRef = ref()

const buildDefaultForm = () => ({
  id: undefined,
  statDate: new Date(),
  newUsers: undefined,
  depositUsers: undefined,
  depositAmount: undefined,
  gameWinloseUsers: undefined,
  retention1: undefined,
  retention3: undefined,
  retention7: undefined,
  retention30: undefined
})

const formData = ref(buildDefaultForm())
const rule = reactive({})

const init = async () => {
  if (route.query.id) {
    const res = await findStatsRetention({ id: route.query.id })
    if (res.code === 0) {
      formData.value = res.data
      type.value = 'update'
      return
    }
  }

  type.value = 'create'
}

const save = async () => {
  btnLoading.value = true
  elFormRef.value?.validate(async (valid) => {
    if (!valid) {
      btnLoading.value = false
      return
    }

    let res
    switch (type.value) {
      case 'update':
        res = await updateStatsRetention(formData.value)
        break
      case 'create':
      default:
        res = await createStatsRetention(formData.value)
        break
    }

    btnLoading.value = false
    if (res.code === 0) {
      ElMessage({
        type: 'success',
        message: '保存成功'
      })
    }
  })
}

const back = () => {
  router.go(-1)
}

init()
</script>
