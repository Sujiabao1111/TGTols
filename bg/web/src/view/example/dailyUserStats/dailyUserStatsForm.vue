
<template>
  <div>
    <div class="gva-form-box">
      <el-form :model="formData" ref="elFormRef" label-position="right" :rules="rule" label-width="80px">
        <el-form-item label="用户ID:" prop="userId">
    <el-input v-model.number="formData.userId" :clearable="true" placeholder="请输入用户ID" />
</el-form-item>
        <el-form-item label="统计日期:" prop="statDate">
    <el-date-picker v-model="formData.statDate" type="date" style="width:100%" placeholder="选择日期" :clearable="true" />
</el-form-item>
        <el-form-item label="当日总流水/打码量:" prop="betAmount">
    <el-input-number v-model="formData.betAmount" style="width:100%" :precision="2" :clearable="true" />
</el-form-item>
        <el-form-item label="当日总输赢(派彩-下注):" prop="winAmount">
    <el-input-number v-model="formData.winAmount" style="width:100%" :precision="2" :clearable="true" />
</el-form-item>
        <el-form-item label="当日总充值:" prop="depositAmount">
    <el-input-number v-model="formData.depositAmount" style="width:100%" :precision="2" :clearable="true" />
</el-form-item>
        <el-form-item label="当日总提现:" prop="withdrawAmount">
    <el-input-number v-model="formData.withdrawAmount" style="width:100%" :precision="2" :clearable="true" />
</el-form-item>
        <el-form-item label="当日登录次数:" prop="loginCount">
    <el-input v-model.number="formData.loginCount" :clearable="true" placeholder="请输入当日登录次数" />
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
import {
  createDailyUserStats,
  updateDailyUserStats,
  findDailyUserStats
} from '@/api/example/dailyUserStats'

defineOptions({
    name: 'DailyUserStatsForm'
})

// 自动获取字典
import { getDictFunc } from '@/utils/format'
import { useRoute, useRouter } from "vue-router"
import { ElMessage } from 'element-plus'
import { ref, reactive } from 'vue'


const route = useRoute()
const router = useRouter()

// 提交按钮loading
const btnLoading = ref(false)

const type = ref('')
const formData = ref({
            userId: undefined,
            statDate: new Date(),
            betAmount: 0,
            winAmount: 0,
            depositAmount: 0,
            withdrawAmount: 0,
            loginCount: undefined,
        })
// 验证规则
const rule = reactive({
})

const elFormRef = ref()

// 初始化方法
const init = async () => {
 // 建议通过url传参获取目标数据ID 调用 find方法进行查询数据操作 从而决定本页面是create还是update 以下为id作为url参数示例
    if (route.query.id) {
      const res = await findDailyUserStats({ ID: route.query.id })
      if (res.code === 0) {
        formData.value = res.data
        type.value = 'update'
      }
    } else {
      type.value = 'create'
    }
}

init()
// 保存按钮
const save = async() => {
      btnLoading.value = true
      elFormRef.value?.validate( async (valid) => {
         if (!valid) return btnLoading.value = false
            let res
           switch (type.value) {
             case 'create':
               res = await createDailyUserStats(formData.value)
               break
             case 'update':
               res = await updateDailyUserStats(formData.value)
               break
             default:
               res = await createDailyUserStats(formData.value)
               break
           }
           btnLoading.value = false
           if (res.code === 0) {
             ElMessage({
               type: 'success',
               message: '创建/更改成功'
             })
           }
       })
}

// 返回按钮
const back = () => {
    router.go(-1)
}

</script>

<style>
</style>
