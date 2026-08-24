
<template>
  <div>
    <div class="gva-search-box">
      <el-form ref="elSearchFormRef" :inline="true" :model="searchInfo" class="demo-form-inline" @keyup.enter="onSubmit">

        <template v-if="showAllQuery">
          <!-- 将需要控制显示状态的查询条件添加到此范围内 -->
        </template>

        <el-form-item>
          <el-button type="primary" icon="search" @click="onSubmit">查询</el-button>
          <el-button icon="refresh" @click="onReset">重置</el-button>
          <el-button link type="primary" icon="arrow-down" @click="showAllQuery=true" v-if="!showAllQuery">展开</el-button>
          <el-button link type="primary" icon="arrow-up" @click="showAllQuery=false" v-else>收起</el-button>
        </el-form-item>
      </el-form>
    </div>
    <div class="gva-table-box">
        <div class="gva-btn-list">
            <el-button  type="primary" icon="plus" @click="openDialog()">新增</el-button>
            <el-button  icon="delete" style="margin-left: 10px;" :disabled="!multipleSelection.length" @click="onDelete">删除</el-button>
            
        </div>
        <el-table
        ref="multipleTable"
        style="width: 100%"
        tooltip-effect="dark"
        :data="tableData"
        row-key="id"
        @selection-change="handleSelectionChange"
        >
        <el-table-column type="selection" width="55" />
        
            <el-table-column align="left" label="id字段" prop="id" width="120" />

            <el-table-column align="left" label="对应前端 key, 如 hero.slide1.title" prop="title" width="120" />

            <el-table-column align="left" label="副标题" prop="subtitle" width="120" />

            <el-table-column align="left" label="标签" prop="tag" width="120" />

            <el-table-column align="left" label="图片URL" prop="image" width="120" />

            <el-table-column align="left" label="Tailwind 渐变色类名" prop="color" width="120" />

            <el-table-column align="left" label="整图跳转链接 (形式2)" prop="jumpLink" width="120" />

            <el-table-column label="按钮配置数组 (形式1)" prop="buttons" width="200">
    <template #default="scope">
        [JSON]
    </template>
</el-table-column>
            <el-table-column align="left" label="排序" prop="sort" width="120" />

            <el-table-column align="left" label="1:启用 0:禁用" prop="status" width="120">
    <template #default="scope">{{ formatBoolean(scope.row.status) }}</template>
</el-table-column>
            <el-table-column align="left" label="createdAt字段" prop="createdAt" width="180">
   <template #default="scope">{{ formatDate(scope.row.createdAt) }}</template>
</el-table-column>
            <el-table-column align="left" label="updatedAt字段" prop="updatedAt" width="180">
   <template #default="scope">{{ formatDate(scope.row.updatedAt) }}</template>
</el-table-column>
        <el-table-column align="left" label="操作" fixed="right" :min-width="appStore.operateMinWith">
            <template #default="scope">
            <el-button  type="primary" link class="table-button" @click="getDetails(scope.row)"><el-icon style="margin-right: 5px"><InfoFilled /></el-icon>查看</el-button>
            <el-button  type="primary" link icon="edit" class="table-button" @click="updateBannersFunc(scope.row)">编辑</el-button>
            <el-button   type="primary" link icon="delete" @click="deleteRow(scope.row)">删除</el-button>
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
    <el-drawer destroy-on-close :size="appStore.drawerSize" v-model="dialogFormVisible" :show-close="false" :before-close="closeDialog">
       <template #header>
              <div class="flex justify-between items-center">
                <span class="text-lg">{{type==='create'?'新增':'编辑'}}</span>
                <div>
                  <el-button :loading="btnLoading" type="primary" @click="enterDialog">确 定</el-button>
                  <el-button @click="closeDialog">取 消</el-button>
                </div>
              </div>
            </template>

          <el-form :model="formData" label-position="top" ref="elFormRef" :rules="rule" label-width="80px">
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
          </el-form>
    </el-drawer>

    <el-drawer destroy-on-close :size="appStore.drawerSize" v-model="detailShow" :show-close="true" :before-close="closeDetailShow" title="查看">
            <el-descriptions :column="1" border>
                    <el-descriptions-item label="id字段">
    {{ detailFrom.id }}
</el-descriptions-item>
                    <el-descriptions-item label="对应前端 key, 如 hero.slide1.title">
    {{ detailFrom.title }}
</el-descriptions-item>
                    <el-descriptions-item label="副标题">
    {{ detailFrom.subtitle }}
</el-descriptions-item>
                    <el-descriptions-item label="标签">
    {{ detailFrom.tag }}
</el-descriptions-item>
                    <el-descriptions-item label="图片URL">
    {{ detailFrom.image }}
</el-descriptions-item>
                    <el-descriptions-item label="Tailwind 渐变色类名">
    {{ detailFrom.color }}
</el-descriptions-item>
                    <el-descriptions-item label="整图跳转链接 (形式2)">
    {{ detailFrom.jumpLink }}
</el-descriptions-item>
                    <el-descriptions-item label="按钮配置数组 (形式1)">
    {{ detailFrom.buttons }}
</el-descriptions-item>
                    <el-descriptions-item label="排序">
    {{ detailFrom.sort }}
</el-descriptions-item>
                    <el-descriptions-item label="1:启用 0:禁用">
    {{ detailFrom.status }}
</el-descriptions-item>
                    <el-descriptions-item label="createdAt字段">
    {{ detailFrom.createdAt }}
</el-descriptions-item>
                    <el-descriptions-item label="updatedAt字段">
    {{ detailFrom.updatedAt }}
</el-descriptions-item>
            </el-descriptions>
        </el-drawer>

  </div>
</template>

<script setup>
import {
  createBanners,
  deleteBanners,
  deleteBannersByIds,
  updateBanners,
  findBanners,
  getBannersList
} from '@/api/example/banners'

// 全量引入格式化工具 请按需保留
import { getDictFunc, formatDate, formatBoolean, filterDict ,filterDataSource, returnArrImg, onDownloadFile } from '@/utils/format'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ref, reactive } from 'vue'
import { useAppStore } from "@/pinia"




defineOptions({
    name: 'Banners'
})

// 提交按钮loading
const btnLoading = ref(false)
const appStore = useAppStore()

// 控制更多查询条件显示/隐藏状态
const showAllQuery = ref(false)

// 自动化生成的字典（可能为空）以及字段
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
const elSearchFormRef = ref()

// =========== 表格控制部分 ===========
const page = ref(1)
const total = ref(0)
const pageSize = ref(10)
const tableData = ref([])
const searchInfo = ref({})
// 重置
const onReset = () => {
  searchInfo.value = {}
  getTableData()
}

// 搜索
const onSubmit = () => {
  elSearchFormRef.value?.validate(async(valid) => {
    if (!valid) return
    page.value = 1
    if (searchInfo.value.status === ""){
        searchInfo.value.status=null
    }
    getTableData()
  })
}

// 分页
const handleSizeChange = (val) => {
  pageSize.value = val
  getTableData()
}

// 修改页面容量
const handleCurrentChange = (val) => {
  page.value = val
  getTableData()
}

// 查询
const getTableData = async() => {
  const table = await getBannersList({ page: page.value, pageSize: pageSize.value, ...searchInfo.value })
  if (table.code === 0) {
    tableData.value = table.data.list
    total.value = table.data.total
    page.value = table.data.page
    pageSize.value = table.data.pageSize
  }
}

getTableData()

// ============== 表格控制部分结束 ===============

// 获取需要的字典 可能为空 按需保留
const setOptions = async () =>{
}

// 获取需要的字典 可能为空 按需保留
setOptions()


// 多选数据
const multipleSelection = ref([])
// 多选
const handleSelectionChange = (val) => {
    multipleSelection.value = val
}

// 删除行
const deleteRow = (row) => {
    ElMessageBox.confirm('确定要删除吗?', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
    }).then(() => {
            deleteBannersFunc(row)
        })
    }

// 多选删除
const onDelete = async() => {
  ElMessageBox.confirm('确定要删除吗?', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async() => {
      const ids = []
      if (multipleSelection.value.length === 0) {
        ElMessage({
          type: 'warning',
          message: '请选择要删除的数据'
        })
        return
      }
      multipleSelection.value &&
        multipleSelection.value.map(item => {
          ids.push(item.id)
        })
      const res = await deleteBannersByIds({ ids })
      if (res.code === 0) {
        ElMessage({
          type: 'success',
          message: '删除成功'
        })
        if (tableData.value.length === ids.length && page.value > 1) {
          page.value--
        }
        getTableData()
      }
      })
    }

// 行为控制标记（弹窗内部需要增还是改）
const type = ref('')

// 更新行
const updateBannersFunc = async(row) => {
    const res = await findBanners({ id: row.id })
    type.value = 'update'
    if (res.code === 0) {
        formData.value = res.data
        dialogFormVisible.value = true
    }
}


// 删除行
const deleteBannersFunc = async (row) => {
    const res = await deleteBanners({ id: row.id })
    if (res.code === 0) {
        ElMessage({
                type: 'success',
                message: '删除成功'
            })
            if (tableData.value.length === 1 && page.value > 1) {
            page.value--
        }
        getTableData()
    }
}

// 弹窗控制标记
const dialogFormVisible = ref(false)

// 打开弹窗
const openDialog = () => {
    type.value = 'create'
    dialogFormVisible.value = true
}

// 关闭弹窗
const closeDialog = () => {
    dialogFormVisible.value = false
    formData.value = {
        title: '',
        subtitle: '',
        tag: '',
        image: '',
        color: '',
        jumpLink: '',
        buttons: {},
        sort: undefined,
        status: false,
        }
}
// 弹窗确定
const enterDialog = async () => {
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
                closeDialog()
                getTableData()
              }
      })
}

const detailFrom = ref({})

// 查看详情控制标记
const detailShow = ref(false)


// 打开详情弹窗
const openDetailShow = () => {
  detailShow.value = true
}


// 打开详情
const getDetails = async (row) => {
  // 打开弹窗
  const res = await findBanners({ id: row.id })
  if (res.code === 0) {
    detailFrom.value = res.data
    openDetailShow()
  }
}


// 关闭详情弹窗
const closeDetailShow = () => {
  detailShow.value = false
  detailFrom.value = {}
}


</script>

<style>

</style>
