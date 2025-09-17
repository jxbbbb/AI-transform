import axios,{ AxiosError } from "axios"
import { ElMessage } from 'element-plus'
import {getCookieValue} from '../utils/utils'

export interface ErrorMsg {
    error?:string
}
const service  = axios.create({
    baseURL: import.meta.env.VITE_API_BASE_URL
})
service.interceptors.request.use(
    function(config){
        //请求之前做点什么，比如添加令牌到请求头
        let access_token = getCookieValue("sso_0voice_access_token")
        if (access_token) {
            config.headers.Authorization = "Bearer " + access_token;
        }else{
            window.location.href = import.meta.env.VITE_USER_CENTER
        }
        return config
    },
    function(error){
        //如果发送错误做点什么
        return Promise.reject(error);
    }
)
service.interceptors.response.use(
    function (response) {
        // 2xx 范围内的状态码都会触发该函数。
        // 对响应数据做点什么
        return response;
      }, function (error) {
        // 超出 2xx 范围的状态码都会触发该函数。
        // 对响应错误做点什么
        const axiosErr = error as AxiosError
        if (!axiosErr.response?.status) {
            ElMessage({
                showClose: true,
                message: axiosErr.message,
                type: 'error',
              })
        }else if(axiosErr.response?.status == 500){
            let message = "服务器内部错误"
            if (axiosErr.response.data) {
                let errorMsg = axiosErr.response.data as ErrorMsg
                if (errorMsg.error) {
                    message = errorMsg.error
                }
            }
            ElMessage({
                showClose: true,
                message:message,
                type: 'error',
              })
        }else if(axiosErr.response?.status == 504){
            ElMessage({
                showClose: true,
                message: "网关超时",
                type: 'error',
              })
        }else if(axiosErr.response?.status == 413){
            ElMessage({
                showClose: true,
                message: "仅支持不大于500M的mp4格式的视频文件",
                type: 'error',
              })
        }else if(axiosErr.response?.status ==400){
            ElMessage({
                showClose: true,
                message: "无效的输入",
                type: 'error',
              })
        }else if (axiosErr.response?.status == 401) {
            window.location.href = import.meta.env.VITE_USER_CENTER
        }
        return Promise.reject(error);
      }   
)
export default service