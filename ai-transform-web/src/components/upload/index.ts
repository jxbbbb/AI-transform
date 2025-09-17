import {get,http} from '../../request/request'

export function cosPresignedUrl<T=any>(filename:string){
    const path = "/v1/cos/presigned/url"
    return get<T>({url:path+"?filename="+filename})
}

export function uploadCos<T=any>(presignedUrl:string,fileContent: ArrayBuffer){
    return http<T>({url:presignedUrl,data:fileContent,method:"PUT"})
}