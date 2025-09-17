<template>
    <div v-loading="loading" class="loading">

        <el-card class="item">
            <el-icon class="avatar-uploader-icon" style="width: 3.125rem; height:3.125rem;">
                <Plus style="width:3.125rem;height:3.125rem" @click="addTranslate" />
            </el-icon>
        </el-card>


        <el-card v-for="r in recordList" class="item">
            <video :src="r.original_video_url" class="item-video"> </video>
                <span class="item-video-footer"></span>

            <el-icon v-if="r.translated_video_url != ''" size="25px" color="green" class="item-check">
                <Check />
            </el-icon>

            <el-icon v-if="r.translated_video_url != ''" size="25px" class="item-word">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 20 20" width="16" height="16">
                    <path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
                        d="m4.166 6.667 4.167 4.166m-5 .834 5-5 1.666-2.5m-8.333 0h10m-5.833-2.5h.833m4.094 12.5h5.978m-5.978 0L9.166 17.5m1.594-3.333 2.388-4.993c.192-.402.289-.603.42-.667a.42.42 0 0 1 .363 0c.131.064.227.265.42.667l2.387 4.993m0 0 1.595 3.333">
                    </path>
                </svg>
            </el-icon>

            <el-icon v-if="r.translated_video_url != ''" size="25px" color="green" class="item-download">
                <Download @click="downloadFile(r.translated_video_url, r.project_name)" />
            </el-icon>

            <div class="item-detail">
                <div class="item-detail-name">{{ r.project_name }}</div>
                <div class="item-detail-time"> {{ getDateStr(r.create_at) }}</div>
            </div>
        </el-card>
    </div>
    <el-dialog v-model="dialogFormVisible" title="Translate" width="500">
        <div v-loading="loading1">
            <el-form :model="form">
                <el-form-item label="Project name" :label-width="formLabelWidth">
                    <el-input v-model="form.project_name" autocomplete="off" />
                </el-form-item>
                <el-form-item label="Upload video" :label-width="formLabelWidth">
                    <el-input autocomplete="off" v-model="form.filename" readonly="false"
                        placeholder="Please select a video" @click="selectFile" />
                </el-form-item>
                <el-form-item label="Original language" :label-width="formLabelWidth">
                    <el-select v-model="form.original_language" placeholder="Please select a language">
                        <el-option label="ZH" value="zh" />
                        <el-option label="EN" value="en" />
                    </el-select>
                </el-form-item>
                <el-form-item label="Translate to" :label-width="formLabelWidth">
                    <el-select v-model="form.translate_language" placeholder="Please select a language">
                        <el-option label="EN" value="en" />
                        <el-option label="ZH" value="zh" />
                    </el-select>
                </el-form-item>
            </el-form>
            <footer class="el-dialog__footer">
                <div class="dialog-footer">
                    <el-button @click="dialogCancel">Cancel</el-button>
                    <el-button type="primary" @click="dialogSubmit">
                        Submit
                    </el-button>
                </div>
            </footer>
        </div>
    </el-dialog>
</template>

<script lang="ts" setup>
import { Plus, Check, Download } from '@element-plus/icons-vue'
import { reactive, ref, onBeforeMount } from "vue"
import { ElNotification } from 'element-plus'
import { record, getTranslateRecords, transInfo, translate } from "./TranslateRecords.ts"
import { getDateStr} from "../../utils/utils"
import { cosPresignedUrl,uploadCos } from "../upload/index.ts"

const loading = ref(false)
const loading1 = ref(false)
let recordList = reactive([] as record[]);

onBeforeMount(() => {
    loadRecords()
})

function loadRecords() {
    loading.value = true
    getTranslateRecords().then(function (res) {
        //清空数组,并重新赋值
        recordList.splice(0);
        for (let a of res.data) {
            let item = {
                id: a.id,
                project_name: a.project_name,
                original_language: a.original_language,
                translated_language: a.translated_language,
                original_video_url: a.original_video_url,
                translated_video_url: a.translated_video_url,
                expiration_at: a.expiration_at,
                create_at: a.create_at,
            };
            recordList.push(item);
        }
        console.log(recordList)
    }).catch((res) => {
        console.log(res);
    }).finally(()=>{
        loading.value = false
    });
}

function addTranslate() {
    dialogFormVisible.value = true
}


const dialogFormVisible = ref(false)
const formLabelWidth = '140px'

let file: File | null;
const form = reactive({
    project_name: '',
    original_language: '',
    translate_language: '',
    filename: '',
    file_url: '',
})

function dialogCancel() {
    dialogFormVisible.value = false
    resetForm()
}
function dialogSubmit() {
    loading1.value = true
    if (file == null) {
        ElNotification({
            title: 'Error',
            message: '请选择一个视频文件',
            type: 'error',
        })
    } else {
        let ok = isVideoFile(file.name)
        if (!ok) {
            ElNotification({
                title: 'Error',
                message: '请选择一个视频文件',
                type: 'error',
            })
            return
        }
    }
    if (form.project_name == "") {
        ElNotification({
            title: 'Error',
            message: '请输入项目名称',
            type: 'error',
        })
        return
    }
    if (form.original_language == "") {
        ElNotification({
            title: 'Error',
            message: '请选择视频语言',
            type: 'error',
        })
        return
    }
    if (form.translate_language == "") {
        ElNotification({
            title: 'Error',
            message: '请选择视频转换的目标语言',
            type: 'error',
        })
        return
    }
    let params: transInfo = {
        project_name: form.project_name,
        original_language: form.original_language,
        translate_language: form.translate_language,
        file_url: form.file_url,
    }
    translate(params).then(() => { }).catch((res) => {
        ElNotification({
            title: 'Error',
            message: res.message,
            type: 'error',
        })

    }).finally(() => {
        loading1.value = false
        resetForm()
        dialogFormVisible.value = false
        loadRecords()
    })
}


function selectFile() {
    const input = document.createElement('input')
    input.type = "file"
    input.id = "file-upload"
    input.addEventListener("change", (event) => {
        const files = (event.target as HTMLInputElement).files
        if (files != null && files.length > 0) {
            let ok = isVideoFile(files[0].name)
            if (!ok) {
                ElNotification({
                    title: 'Error',
                    message: '仅支持mp4格式的视频文件',
                    type: 'error',
                })
                return
            }
            file = files[0]

            loading1.value = true
            //获取预签名链接
            cosPresignedUrl(file.name).then((res) => {
               let presignedUrl = res.data?.presigned_url
               let file_url = res.data?.file_url

                file?.arrayBuffer().then((content) => {
                    //web直传cos
                    uploadCos(presignedUrl, content).then(() => {
                        form.file_url = file_url 
                        if (file?.name != null){
                            form.filename = file.name
                        }
                    }).catch((err) => {
                        ElNotification({
                            title: 'Error',
                            message: err.message,
                            type: 'error',
                        })
                        return
                    }).finally(()=>{
                        loading1.value = false
                    })
                })
            })
        }
    });
    input.click()
}
function isVideoFile(fileName: string): boolean {
    return fileName.endsWith(".mp4") ||
        fileName.endsWith(".mov") ||
        fileName.endsWith(".avi") ||
        fileName.endsWith(".mkv") ||
        fileName.endsWith(".flv");
}

function resetForm() {
    form.original_language = ""
    form.project_name = ""
    form.translate_language = ""
    file = null
}

function downloadFile(url: string, fileName: string) {
    fetch(url)
        .then(response => response.blob())
        .then(blob => {
            const a = document.createElement('a');
            const url = URL.createObjectURL(blob);
            a.href = url;
            a.download = fileName;

            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            URL.revokeObjectURL(url);
        });
}

</script>
<style>
.loading {
    position: relative;
    display: flex;
    white-space: normal;
    word-wrap: break-word;
    width: 100%;
    height: 100%;
    flex-wrap: wrap;
    align-content: first baseline;
    justify-content: center;
}
.item {
    width: 22rem;
    height: 15.75rem;
    border-radius: 1.25rem;
    margin-right: 0.5rem;
    margin-bottom: 1rem;
    padding: 0;
    flex-wrap: wrap;
    position: relative;
    font-family: Inter;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-wrap: wrap;
}
.item:hover {
    border: 2px solid rgb(15, 7, 243);
}
.item-video {
    width: 100%;
    height: 11.625rem;
    border-radius: 1.25rem;
    position:absolute;
    top:0;
    left: 0;
}
.item-video-footer {
    width: 100%;
    height: 0.3rem;
    background-color: blue;
    position:absolute;
    border-bottom-right-radius: 1.25rem;
    border-bottom-left-radius: 1.25rem;
    top:11.325rem;
    left: 0;
}
.item-check {
    font-weight: bold;
    position: absolute;
    left: 1rem;
    top: 9.5rem;
    background-color: white;
    border-radius: 25px;
}
.item-word {
    font-weight: bold;
    position: absolute;
    left: 1rem;
    top: 0.5rem;
    background-color: white;
    border-radius: 25px;
}
.item-download {
    font-weight: bold;
    position: absolute;
    right: 0.5rem;
    top: 0.5rem;
    background-color: white;
    border-radius: 25px;
}
.item-detail {
    width: 22rem;
    height: 4.125rem;
    left: 0;
    top: 11.625rem;
    position: absolute;
}
.item-detail-name {
    font-weight: bold;
    position: absolute;
    left: 1rem;
    top: .2rem;
}
.item-detail-time {
    position: absolute;
    left: 1rem;
    top: 2rem;
    color: rgb(102, 112, 133);
}
.el-card__body {
    padding: 0;
}

.show {
    display: block;
}

.hidden {
    display: none;
}
</style>