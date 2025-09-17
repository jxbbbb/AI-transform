import {get,post} from '../../request/request'

export interface record {
    id:number,
    project_name:string,
    original_language:string,
    translated_language:string,
    original_video_url:string,
    translated_video_url:string,
    expiration_at:number,
    create_at:number,
}

export function getTranslateRecords<T=any>(){
    const path = "/v1/records"
    return get<T>({url:path})
}


export interface transInfo {
    project_name:string,
    original_language:string,
    translate_language:string,    
    file_url:string,
}

export function translate<T=any>(params:transInfo){
    const path = "/v1/translate"
    let formData = new FormData()
    formData.append("project_name",params.project_name)
    formData.append("original_language",params.original_language)
    formData.append("translate_language",params.translate_language)
    formData.append("file_url",params.file_url)
    return post<T>({url:path,data:formData})
}
