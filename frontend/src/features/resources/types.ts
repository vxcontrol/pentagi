export interface ResourceItem extends ResourcePreviewModel {
    isUploaded: boolean;
    progress: number;
}

export interface ResourcePreviewModel {
    format: string;
    id: string;
    name: string;
    uploadedAt: string;
}

export interface UploadUrlResponse {
    id: string;
    item: ResourcePreviewModel;
    uploadUrl: string;
}
