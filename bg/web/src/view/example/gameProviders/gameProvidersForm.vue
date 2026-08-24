
<template>
  <div>
    <div class="gva-form-box">
      <el-form :model="formData" ref="elFormRef" label-position="right" :rules="rule" label-width="80px">
        <el-form-item label="厂商代码 (如 PG, AG):" prop="code">
    <el-input v-model="formData.code" :clearable="true" placeholder="请输入厂商代码 (如 PG, AG)" />
</el-form-item>
        <el-form-item label="厂商显示名称:" prop="name">
    <el-input v-model="formData.name" :clearable="true" placeholder="请输入厂商显示名称" />
</el-form-item>
        <el-form-item label="JSON格式存储API Key, Secret, Endpoint等敏感信息:" prop="apiConfig">
    // 此字段为json结构，可以前端自行控制展示和数据绑定模式 需绑定json的key为 formData.apiConfig 后端会按照json的类型进行存取
    {{ formData.apiConfig }}
</el-form-item>
        <el-form-item label="1:开启 0:维护:" prop="status">
    <el-switch v-model="formData.status" active-color="#13ce66" inactive-color="#ff4949" active-text="是" inactive-text="否" clearable ></el-switch>
</el-form-item>
        <el-form-item label="createdAt字段:" prop="createdAt">
    <el-date-picker v-model="formData.createdAt" type="date" style="width:100%" placeholder="选择日期" :clearable="true" />
</el-form-item>
        <el-form-item label="游戏类型Id:" prop="typeId">
    <el-input v-model.number="formData.typeId" :clearable="true" placeholder="请输入游戏类型Id" />
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
  createGameProviders,
  updateGameProviders,
  findGameProviders
} from '@/api/example/gameProviders'

defineOptions({
    name: 'GameProvidersForm'
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
            code: '',
            name: '',
            apiConfig: {},
            status: false,
            createdAt: new Date(),
            typeId: undefined,
        })
// 验证规则
const rule = reactive({
})

const elFormRef = ref()

// 初始化方法
const init = async () => {
 // 建议通过url传参获取目标数据ID 调用 find方法进行查询数据操作 从而决定本页面是create还是update 以下为id作为url参数示例
    if (route.query.id) {
      const res = await findGameProviders({ ID: route.query.id })
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
               res = await createGameProviders(formData.value)
               break
             case 'update':
               res = await updateGameProviders(formData.value)
               break
             default:
               res = await createGameProviders(formData.value)
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
