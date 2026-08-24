
<template>
  <div>
    <div class="gva-form-box">
      <el-form :model="formData" ref="elFormRef" label-position="right" :rules="rule" label-width="80px">
        <el-form-item label="对应前端 key, 如 hero.slide1.title:" prop="title">
    <el-input v-model="formData.title" :clearable="true" placeholder="请输入对应前端 key, 如 hero.slide1.title" />
</el-form-item>
        <el-form-item label="副标题:" prop="subtitle">
    <el-input v-model="formData.subtitle" :clearable="true" placeholder="请输入副标题" />
</el-form-item>
        <el-form-item label="标签:" prop="tag">
    <el-input v-model="formData.tag" :clearable="true" placeholder="请输入标签" />
</el-form-item>
        <el-form-item label="图片URL:" prop="image">
    <el-input v-model="formData.image" :clearable="true" placeholder="请输入图片URL" />
</el-form-item>
        <el-form-item label="Tailwind 渐变色类名:" prop="color">
    <el-input v-model="formData.color" :clearable="true" placeholder="请输入Tailwind 渐变色类名" />
</el-form-item>
        <el-form-item label="整图跳转链接 (形式2):" prop="jumpLink">
    <el-input v-model="formData.jumpLink" :clearable="true" placeholder="请输入整图跳转链接 (形式2)" />
</el-form-item>
        <el-form-item label="按钮配置数组 (形式1):" prop="buttons">
    // 此字段为json结构，可以前端自行控制展示和数据绑定模式 需绑定json的key为 formData.buttons 后端会按照json的类型进行存取
    {{ formData.buttons }}
</el-form-item>
        <el-form-item label="排序:" prop="sort">
    <el-input v-model.number="formData.sort" :clearable="true" placeholder="请输入排序" />
</el-form-item>
        <el-form-item label="1:启用 0:禁用:" prop="status">
    <el-switch v-model="formData.status" active-color="#13ce66" inactive-color="#ff4949" active-text="是" inactive-text="否" clearable ></el-switch>
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
  createBanners,
  updateBanners,
  findBanners
} from '@/api/example/banners'

defineOptions({
    name: 'BannersForm'
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
            title: '',
            subtitle: '',
            tag: '',
            image: '',
            color: '',
            jumpLink: '',
            buttons: {},
            sort: undefined,
            status: false,
        })
// 验证规则
const rule = reactive({
})

const elFormRef = ref()

// 初始化方法
const init = async () => {
 // 建议通过url传参获取目标数据ID 调用 find方法进行查询数据操作 从而决定本页面是create还是update 以下为id作为url参数示例
    if (route.query.id) {
      const res = await findBanners({ ID: route.query.id })
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
               res = await createBanners(formData.value)
               break
             case 'update':
               res = await updateBanners(formData.value)
               break
             default:
               res = await createBanners(formData.value)
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
