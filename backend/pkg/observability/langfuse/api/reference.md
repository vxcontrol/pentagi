# Reference
## Annotationqueues
<details><summary><code>client.Annotationqueues.Listqueues() -> *api.PaginatedAnnotationQueues</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Get all annotation queues
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.AnnotationQueuesListQueuesRequest{}
client.Annotationqueues.Listqueues(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**page:** `*int` — page number, starts at 1
    
</dd>
</dl>

<dl>
<dd>

**limit:** `*int` — limit of items per page
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Annotationqueues.Createqueue(request) -> *api.AnnotationQueue</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Create an annotation queue
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.CreateAnnotationQueueRequest{
        Name: "name",
        ScoreConfigIDs: []string{
            "scoreConfigIds",
        },
    }
client.Annotationqueues.Createqueue(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**name:** `string` 
    
</dd>
</dl>

<dl>
<dd>

**description:** `*string` 
    
</dd>
</dl>

<dl>
<dd>

**scoreConfigIDs:** `[]string` 
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Annotationqueues.Getqueue(QueueID) -> *api.AnnotationQueue</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Get an annotation queue by ID
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.AnnotationQueuesGetQueueRequest{
        QueueID: "queueId",
    }
client.Annotationqueues.Getqueue(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**queueID:** `string` — The unique identifier of the annotation queue
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Annotationqueues.Listqueueitems(QueueID) -> *api.PaginatedAnnotationQueueItems</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Get items for a specific annotation queue
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.AnnotationQueuesListQueueItemsRequest{
        QueueID: "queueId",
    }
client.Annotationqueues.Listqueueitems(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**queueID:** `string` — The unique identifier of the annotation queue
    
</dd>
</dl>

<dl>
<dd>

**status:** `*api.AnnotationQueueStatus` — Filter by status
    
</dd>
</dl>

<dl>
<dd>

**page:** `*int` — page number, starts at 1
    
</dd>
</dl>

<dl>
<dd>

**limit:** `*int` — limit of items per page
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Annotationqueues.Createqueueitem(QueueID, request) -> *api.AnnotationQueueItem</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Add an item to an annotation queue
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.CreateAnnotationQueueItemRequest{
        QueueID: "queueId",
        ObjectID: "objectId",
        ObjectType: api.AnnotationQueueObjectTypeTrace,
    }
client.Annotationqueues.Createqueueitem(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**queueID:** `string` — The unique identifier of the annotation queue
    
</dd>
</dl>

<dl>
<dd>

**objectID:** `string` 
    
</dd>
</dl>

<dl>
<dd>

**objectType:** `*api.AnnotationQueueObjectType` 
    
</dd>
</dl>

<dl>
<dd>

**status:** `*api.AnnotationQueueStatus` — Defaults to PENDING for new queue items
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Annotationqueues.Getqueueitem(QueueID, ItemID) -> *api.AnnotationQueueItem</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Get a specific item from an annotation queue
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.AnnotationQueuesGetQueueItemRequest{
        QueueID: "queueId",
        ItemID: "itemId",
    }
client.Annotationqueues.Getqueueitem(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**queueID:** `string` — The unique identifier of the annotation queue
    
</dd>
</dl>

<dl>
<dd>

**itemID:** `string` — The unique identifier of the annotation queue item
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Annotationqueues.Deletequeueitem(QueueID, ItemID) -> *api.DeleteAnnotationQueueItemResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Remove an item from an annotation queue
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.AnnotationQueuesDeleteQueueItemRequest{
        QueueID: "queueId",
        ItemID: "itemId",
    }
client.Annotationqueues.Deletequeueitem(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**queueID:** `string` — The unique identifier of the annotation queue
    
</dd>
</dl>

<dl>
<dd>

**itemID:** `string` — The unique identifier of the annotation queue item
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Annotationqueues.Updatequeueitem(QueueID, ItemID, request) -> *api.AnnotationQueueItem</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Update an annotation queue item
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.UpdateAnnotationQueueItemRequest{
        QueueID: "queueId",
        ItemID: "itemId",
    }
client.Annotationqueues.Updatequeueitem(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**queueID:** `string` — The unique identifier of the annotation queue
    
</dd>
</dl>

<dl>
<dd>

**itemID:** `string` — The unique identifier of the annotation queue item
    
</dd>
</dl>

<dl>
<dd>

**status:** `*api.AnnotationQueueStatus` 
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Annotationqueues.Createqueueassignment(QueueID, request) -> *api.CreateAnnotationQueueAssignmentResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Create an assignment for a user to an annotation queue
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.AnnotationQueuesCreateQueueAssignmentRequest{
        QueueID: "queueId",
        Body: &api.AnnotationQueueAssignmentRequest{
            UserID: "userId",
        },
    }
client.Annotationqueues.Createqueueassignment(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**queueID:** `string` — The unique identifier of the annotation queue
    
</dd>
</dl>

<dl>
<dd>

**request:** `*api.AnnotationQueueAssignmentRequest` 
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Annotationqueues.Deletequeueassignment(QueueID, request) -> *api.DeleteAnnotationQueueAssignmentResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Delete an assignment for a user to an annotation queue
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.AnnotationQueuesDeleteQueueAssignmentRequest{
        QueueID: "queueId",
        Body: &api.AnnotationQueueAssignmentRequest{
            UserID: "userId",
        },
    }
client.Annotationqueues.Deletequeueassignment(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**queueID:** `string` — The unique identifier of the annotation queue
    
</dd>
</dl>

<dl>
<dd>

**request:** `*api.AnnotationQueueAssignmentRequest` 
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

## Blobstorageintegrations
<details><summary><code>client.Blobstorageintegrations.Getblobstorageintegrations() -> *api.BlobStorageIntegrationsResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Get all blob storage integrations for the organization (requires organization-scoped API key)
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
client.Blobstorageintegrations.Getblobstorageintegrations(
        context.TODO(),
    )
}
```
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Blobstorageintegrations.Upsertblobstorageintegration(request) -> *api.BlobStorageIntegrationResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Create or update a blob storage integration for a specific project (requires organization-scoped API key). The configuration is validated by performing a test upload to the bucket.
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.CreateBlobStorageIntegrationRequest{
        ProjectID: "projectId",
        Type: api.BlobStorageIntegrationTypeS3,
        BucketName: "bucketName",
        Region: "region",
        ExportFrequency: api.BlobStorageExportFrequencyHourly,
        Enabled: true,
        ForcePathStyle: true,
        FileType: api.BlobStorageIntegrationFileTypeJSON,
        ExportMode: api.BlobStorageExportModeFullHistory,
    }
client.Blobstorageintegrations.Upsertblobstorageintegration(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**projectID:** `string` — ID of the project in which to configure the blob storage integration
    
</dd>
</dl>

<dl>
<dd>

**type_:** `*api.BlobStorageIntegrationType` 
    
</dd>
</dl>

<dl>
<dd>

**bucketName:** `string` — Name of the storage bucket
    
</dd>
</dl>

<dl>
<dd>

**endpoint:** `*string` — Custom endpoint URL (required for S3_COMPATIBLE type)
    
</dd>
</dl>

<dl>
<dd>

**region:** `string` — Storage region
    
</dd>
</dl>

<dl>
<dd>

**accessKeyID:** `*string` — Access key ID for authentication
    
</dd>
</dl>

<dl>
<dd>

**secretAccessKey:** `*string` — Secret access key for authentication (will be encrypted when stored)
    
</dd>
</dl>

<dl>
<dd>

**prefix:** `*string` — Path prefix for exported files (must end with forward slash if provided)
    
</dd>
</dl>

<dl>
<dd>

**exportFrequency:** `*api.BlobStorageExportFrequency` 
    
</dd>
</dl>

<dl>
<dd>

**enabled:** `bool` — Whether the integration is active
    
</dd>
</dl>

<dl>
<dd>

**forcePathStyle:** `bool` — Use path-style URLs for S3 requests
    
</dd>
</dl>

<dl>
<dd>

**fileType:** `*api.BlobStorageIntegrationFileType` 
    
</dd>
</dl>

<dl>
<dd>

**exportMode:** `*api.BlobStorageExportMode` 
    
</dd>
</dl>

<dl>
<dd>

**exportStartDate:** `*time.Time` — Custom start date for exports (required when exportMode is FROM_CUSTOM_DATE)
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Blobstorageintegrations.Deleteblobstorageintegration(ID) -> *api.BlobStorageIntegrationDeletionResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Delete a blob storage integration by ID (requires organization-scoped API key)
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.BlobStorageIntegrationsDeleteBlobStorageIntegrationRequest{
        ID: "id",
    }
client.Blobstorageintegrations.Deleteblobstorageintegration(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**id:** `string` 
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

## Comments
<details><summary><code>client.Comments.Get() -> *api.GetCommentsResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Get all comments
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.CommentsGetRequest{}
client.Comments.Get(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**page:** `*int` — Page number, starts at 1.
    
</dd>
</dl>

<dl>
<dd>

**limit:** `*int` — Limit of items per page. If you encounter api issues due to too large page sizes, try to reduce the limit
    
</dd>
</dl>

<dl>
<dd>

**objectType:** `*string` — Filter comments by object type (trace, observation, session, prompt).
    
</dd>
</dl>

<dl>
<dd>

**objectID:** `*string` — Filter comments by object id. If objectType is not provided, an error will be thrown.
    
</dd>
</dl>

<dl>
<dd>

**authorUserID:** `*string` — Filter comments by author user id.
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Comments.Create(request) -> *api.CreateCommentResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Create a comment. Comments may be attached to different object types (trace, observation, session, prompt).
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.CreateCommentRequest{
        ProjectID: "projectId",
        ObjectType: "objectType",
        ObjectID: "objectId",
        Content: "content",
    }
client.Comments.Create(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**projectID:** `string` — The id of the project to attach the comment to.
    
</dd>
</dl>

<dl>
<dd>

**objectType:** `string` — The type of the object to attach the comment to (trace, observation, session, prompt).
    
</dd>
</dl>

<dl>
<dd>

**objectID:** `string` — The id of the object to attach the comment to. If this does not reference a valid existing object, an error will be thrown.
    
</dd>
</dl>

<dl>
<dd>

**content:** `string` — The content of the comment. May include markdown. Currently limited to 5000 characters.
    
</dd>
</dl>

<dl>
<dd>

**authorUserID:** `*string` — The id of the user who created the comment.
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Comments.GetByID(CommentID) -> *api.Comment</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Get a comment by id
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.CommentsGetByIDRequest{
        CommentID: "commentId",
    }
client.Comments.GetByID(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**commentID:** `string` — The unique langfuse identifier of a comment
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

## Datasetitems
<details><summary><code>client.Datasetitems.List() -> *api.PaginatedDatasetItems</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Get dataset items
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.DatasetItemsListRequest{}
client.Datasetitems.List(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**datasetName:** `*string` 
    
</dd>
</dl>

<dl>
<dd>

**sourceTraceID:** `*string` 
    
</dd>
</dl>

<dl>
<dd>

**sourceObservationID:** `*string` 
    
</dd>
</dl>

<dl>
<dd>

**page:** `*int` — page number, starts at 1
    
</dd>
</dl>

<dl>
<dd>

**limit:** `*int` — limit of items per page
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Datasetitems.Create(request) -> *api.DatasetItem</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Create a dataset item
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.CreateDatasetItemRequest{
        DatasetName: "datasetName",
    }
client.Datasetitems.Create(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**datasetName:** `string` 
    
</dd>
</dl>

<dl>
<dd>

**input:** `any` 
    
</dd>
</dl>

<dl>
<dd>

**expectedOutput:** `any` 
    
</dd>
</dl>

<dl>
<dd>

**metadata:** `any` 
    
</dd>
</dl>

<dl>
<dd>

**sourceTraceID:** `*string` 
    
</dd>
</dl>

<dl>
<dd>

**sourceObservationID:** `*string` 
    
</dd>
</dl>

<dl>
<dd>

**id:** `*string` — Dataset items are upserted on their id. Id needs to be unique (project-level) and cannot be reused across datasets.
    
</dd>
</dl>

<dl>
<dd>

**status:** `*api.DatasetStatus` — Defaults to ACTIVE for newly created items
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Datasetitems.Get(ID) -> *api.DatasetItem</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Get a dataset item
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.DatasetItemsGetRequest{
        ID: "id",
    }
client.Datasetitems.Get(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**id:** `string` 
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Datasetitems.Delete(ID) -> *api.DeleteDatasetItemResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Delete a dataset item and all its run items. This action is irreversible.
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.DatasetItemsDeleteRequest{
        ID: "id",
    }
client.Datasetitems.Delete(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**id:** `string` 
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

## Datasetrunitems
<details><summary><code>client.Datasetrunitems.List() -> *api.PaginatedDatasetRunItems</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

List dataset run items
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.DatasetRunItemsListRequest{
        DatasetID: "datasetId",
        RunName: "runName",
    }
client.Datasetrunitems.List(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**datasetID:** `string` 
    
</dd>
</dl>

<dl>
<dd>

**runName:** `string` 
    
</dd>
</dl>

<dl>
<dd>

**page:** `*int` — page number, starts at 1
    
</dd>
</dl>

<dl>
<dd>

**limit:** `*int` — limit of items per page
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Datasetrunitems.Create(request) -> *api.DatasetRunItem</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Create a dataset run item
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.CreateDatasetRunItemRequest{
        RunName: "runName",
        DatasetItemID: "datasetItemId",
    }
client.Datasetrunitems.Create(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**runName:** `string` 
    
</dd>
</dl>

<dl>
<dd>

**runDescription:** `*string` — Description of the run. If run exists, description will be updated.
    
</dd>
</dl>

<dl>
<dd>

**metadata:** `any` — Metadata of the dataset run, updates run if run already exists
    
</dd>
</dl>

<dl>
<dd>

**datasetItemID:** `string` 
    
</dd>
</dl>

<dl>
<dd>

**observationID:** `*string` 
    
</dd>
</dl>

<dl>
<dd>

**traceID:** `*string` — traceId should always be provided. For compatibility with older SDK versions it can also be inferred from the provided observationId.
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

## Datasets
<details><summary><code>client.Datasets.List() -> *api.PaginatedDatasets</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Get all datasets
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.DatasetsListRequest{}
client.Datasets.List(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**page:** `*int` — page number, starts at 1
    
</dd>
</dl>

<dl>
<dd>

**limit:** `*int` — limit of items per page
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Datasets.Create(request) -> *api.Dataset</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Create a dataset
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.CreateDatasetRequest{
        Name: "name",
    }
client.Datasets.Create(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**name:** `string` 
    
</dd>
</dl>

<dl>
<dd>

**description:** `*string` 
    
</dd>
</dl>

<dl>
<dd>

**metadata:** `any` 
    
</dd>
</dl>

<dl>
<dd>

**inputSchema:** `any` — JSON Schema for validating dataset item inputs. When set, all new and existing dataset items will be validated against this schema.
    
</dd>
</dl>

<dl>
<dd>

**expectedOutputSchema:** `any` — JSON Schema for validating dataset item expected outputs. When set, all new and existing dataset items will be validated against this schema.
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Datasets.Get(DatasetName) -> *api.Dataset</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Get a dataset
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.DatasetsGetRequest{
        DatasetName: "datasetName",
    }
client.Datasets.Get(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**datasetName:** `string` 
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Datasets.Getrun(DatasetName, RunName) -> *api.DatasetRunWithItems</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Get a dataset run and its items
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.DatasetsGetRunRequest{
        DatasetName: "datasetName",
        RunName: "runName",
    }
client.Datasets.Getrun(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**datasetName:** `string` 
    
</dd>
</dl>

<dl>
<dd>

**runName:** `string` 
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Datasets.Deleterun(DatasetName, RunName) -> *api.DeleteDatasetRunResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Delete a dataset run and all its run items. This action is irreversible.
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.DatasetsDeleteRunRequest{
        DatasetName: "datasetName",
        RunName: "runName",
    }
client.Datasets.Deleterun(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**datasetName:** `string` 
    
</dd>
</dl>

<dl>
<dd>

**runName:** `string` 
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Datasets.Getruns(DatasetName) -> *api.PaginatedDatasetRuns</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Get dataset runs
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.DatasetsGetRunsRequest{
        DatasetName: "datasetName",
    }
client.Datasets.Getruns(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**datasetName:** `string` 
    
</dd>
</dl>

<dl>
<dd>

**page:** `*int` — page number, starts at 1
    
</dd>
</dl>

<dl>
<dd>

**limit:** `*int` — limit of items per page
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

## Health
<details><summary><code>client.Health.Health() -> *api.HealthResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Check health of API and database
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
client.Health.Health(
        context.TODO(),
    )
}
```
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

## Ingestion
<details><summary><code>client.Ingestion.Batch(request) -> *api.IngestionResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

**Legacy endpoint for batch ingestion for Langfuse Observability.**

-> Please use the OpenTelemetry endpoint (`/api/public/otel/v1/traces`). Learn more: https://langfuse.com/integrations/native/opentelemetry

Within each batch, there can be multiple events.
Each event has a type, an id, a timestamp, metadata and a body.
Internally, we refer to this as the "event envelope" as it tells us something about the event but not the trace.
We use the event id within this envelope to deduplicate messages to avoid processing the same event twice, i.e. the event id should be unique per request.
The event.body.id is the ID of the actual trace and will be used for updates and will be visible within the Langfuse App.
I.e. if you want to update a trace, you'd use the same body id, but separate event IDs.

Notes:
- Introduction to data model: https://langfuse.com/docs/observability/data-model
- Batch sizes are limited to 3.5 MB in total. You need to adjust the number of events per batch accordingly.
- The API does not return a 4xx status code for input errors. Instead, it responds with a 207 status code, which includes a list of the encountered errors.
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.IngestionBatchRequest{
        Batch: []*api.IngestionEvent{
            &api.IngestionEvent{
                IngestionEventZero: &api.IngestionEventZero{
                    ID: "abcdef-1234-5678-90ab",
                    Timestamp: "2022-01-01T00:00:00.000Z",
                    Body: &api.TraceBody{
                        ID: api.String(
                            "abcdef-1234-5678-90ab",
                        ),
                        Timestamp: api.Time(
                            api.MustParseDateTime(
                                "2022-01-01T00:00:00Z",
                            ),
                        ),
                        Name: api.String(
                            "My Trace",
                        ),
                        UserID: api.String(
                            "1234-5678-90ab-cdef",
                        ),
                        Input: "My input",
                        Output: "My output",
                        SessionID: api.String(
                            "1234-5678-90ab-cdef",
                        ),
                        Release: api.String(
                            "1.0.0",
                        ),
                        Version: api.String(
                            "1.0.0",
                        ),
                        Metadata: "My metadata",
                        Tags: []string{
                            "tag1",
                            "tag2",
                        },
                        Environment: api.String(
                            "production",
                        ),
                        Public: api.Bool(
                            true,
                        ),
                    },
                    Type: api.IngestionEventZeroTypeTraceCreate.Ptr(),
                },
            },
        },
    }
client.Ingestion.Batch(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**batch:** `[]*api.IngestionEvent` — Batch of tracing events to be ingested. Discriminated by attribute `type`.
    
</dd>
</dl>

<dl>
<dd>

**metadata:** `any` — Optional. Metadata field used by the Langfuse SDKs for debugging.
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

## Llmconnections
<details><summary><code>client.Llmconnections.List() -> *api.PaginatedLlmConnections</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Get all LLM connections in a project
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.LlmConnectionsListRequest{}
client.Llmconnections.List(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**page:** `*int` — page number, starts at 1
    
</dd>
</dl>

<dl>
<dd>

**limit:** `*int` — limit of items per page
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Llmconnections.Upsert(request) -> *api.LlmConnection</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Create or update an LLM connection. The connection is upserted on provider.
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.UpsertLlmConnectionRequest{
        Provider: "provider",
        Adapter: api.LlmAdapterAnthropic,
        SecretKey: "secretKey",
    }
client.Llmconnections.Upsert(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**provider:** `string` — Provider name (e.g., 'openai', 'my-gateway'). Must be unique in project, used for upserting.
    
</dd>
</dl>

<dl>
<dd>

**adapter:** `*api.LlmAdapter` — The adapter used to interface with the LLM
    
</dd>
</dl>

<dl>
<dd>

**secretKey:** `string` — Secret key for the LLM API.
    
</dd>
</dl>

<dl>
<dd>

**baseURL:** `*string` — Custom base URL for the LLM API
    
</dd>
</dl>

<dl>
<dd>

**customModels:** `[]string` — List of custom model names
    
</dd>
</dl>

<dl>
<dd>

**withDefaultModels:** `*bool` — Whether to include default models. Default is true.
    
</dd>
</dl>

<dl>
<dd>

**extraHeaders:** `map[string]*string` — Extra headers to send with requests
    
</dd>
</dl>

<dl>
<dd>

**config:** `map[string]any` — Adapter-specific configuration. Validation rules: - **Bedrock**: Required. Must be `{"region": "<aws-region>"}` (e.g., `{"region":"us-east-1"}`) - **VertexAI**: Optional. If provided, must be `{"location": "<gcp-location>"}` (e.g., `{"location":"us-central1"}`) - **Other adapters**: Not supported. Omit this field or set to null.
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

## Media
<details><summary><code>client.Media.Get(MediaID) -> *api.GetMediaResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Get a media record
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.MediaGetRequest{
        MediaID: "mediaId",
    }
client.Media.Get(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**mediaID:** `string` — The unique langfuse identifier of a media record
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Media.Patch(MediaID, request) -> error</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Patch a media record
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.PatchMediaBody{
        MediaID: "mediaId",
        UploadedAt: api.MustParseDateTime(
            "2024-01-15T09:30:00Z",
        ),
        UploadHTTPStatus: 1,
    }
client.Media.Patch(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**mediaID:** `string` — The unique langfuse identifier of a media record
    
</dd>
</dl>

<dl>
<dd>

**uploadedAt:** `time.Time` — The date and time when the media record was uploaded
    
</dd>
</dl>

<dl>
<dd>

**uploadHTTPStatus:** `int` — The HTTP status code of the upload
    
</dd>
</dl>

<dl>
<dd>

**uploadHTTPError:** `*string` — The HTTP error message of the upload
    
</dd>
</dl>

<dl>
<dd>

**uploadTimeMs:** `*int` — The time in milliseconds it took to upload the media record
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Media.Getuploadurl(request) -> *api.GetMediaUploadURLResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Get a presigned upload URL for a media record
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.GetMediaUploadURLRequest{
        TraceID: "traceId",
        ContentType: api.MediaContentTypeImagePng,
        ContentLength: 1,
        Sha256Hash: "sha256Hash",
        Field: "field",
    }
client.Media.Getuploadurl(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**traceID:** `string` — The trace ID associated with the media record
    
</dd>
</dl>

<dl>
<dd>

**observationID:** `*string` — The observation ID associated with the media record. If the media record is associated directly with a trace, this will be null.
    
</dd>
</dl>

<dl>
<dd>

**contentType:** `*api.MediaContentType` 
    
</dd>
</dl>

<dl>
<dd>

**contentLength:** `int` — The size of the media record in bytes
    
</dd>
</dl>

<dl>
<dd>

**sha256Hash:** `string` — The SHA-256 hash of the media record
    
</dd>
</dl>

<dl>
<dd>

**field:** `string` — The trace / observation field the media record is associated with. This can be one of `input`, `output`, `metadata`
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

## Metricsv2
<details><summary><code>client.Metricsv2.Metrics() -> *api.MetricsV2Response</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Get metrics from the Langfuse project using a query object. V2 endpoint with optimized performance.

## V2 Differences
- Supports `observations`, `scores-numeric`, and `scores-categorical` views only (traces view not supported)
- Direct access to tags and release fields on observations
- Backwards-compatible: traceName, traceRelease, traceVersion dimensions are still available on observations view
- High cardinality dimensions are not supported and will return a 400 error (see below)

For more details, see the [Metrics API documentation](https://langfuse.com/docs/metrics/features/metrics-api).

## Available Views

### observations
Query observation-level data (spans, generations, events).

**Dimensions:**
- `environment` - Deployment environment (e.g., production, staging)
- `type` - Type of observation (SPAN, GENERATION, EVENT)
- `name` - Name of the observation
- `level` - Logging level of the observation
- `version` - Version of the observation
- `tags` - User-defined tags
- `release` - Release version
- `traceName` - Name of the parent trace (backwards-compatible)
- `traceRelease` - Release version of the parent trace (backwards-compatible, maps to release)
- `traceVersion` - Version of the parent trace (backwards-compatible, maps to version)
- `providedModelName` - Name of the model used
- `promptName` - Name of the prompt used
- `promptVersion` - Version of the prompt used
- `startTimeMonth` - Month of start_time in YYYY-MM format

**Measures:**
- `count` - Total number of observations
- `latency` - Observation latency (milliseconds)
- `streamingLatency` - Generation latency from completion start to end (milliseconds)
- `inputTokens` - Sum of input tokens consumed
- `outputTokens` - Sum of output tokens produced
- `totalTokens` - Sum of all tokens consumed
- `outputTokensPerSecond` - Output tokens per second
- `tokensPerSecond` - Total tokens per second
- `inputCost` - Input cost (USD)
- `outputCost` - Output cost (USD)
- `totalCost` - Total cost (USD)
- `timeToFirstToken` - Time to first token (milliseconds)
- `countScores` - Number of scores attached to the observation

### scores-numeric
Query numeric and boolean score data.

**Dimensions:**
- `environment` - Deployment environment
- `name` - Name of the score (e.g., accuracy, toxicity)
- `source` - Origin of the score (API, ANNOTATION, EVAL)
- `dataType` - Data type (NUMERIC, BOOLEAN)
- `configId` - Identifier of the score config
- `timestampMonth` - Month in YYYY-MM format
- `timestampDay` - Day in YYYY-MM-DD format
- `value` - Numeric value of the score
- `traceName` - Name of the parent trace
- `tags` - Tags
- `traceRelease` - Release version
- `traceVersion` - Version
- `observationName` - Name of the associated observation
- `observationModelName` - Model name of the associated observation
- `observationPromptName` - Prompt name of the associated observation
- `observationPromptVersion` - Prompt version of the associated observation

**Measures:**
- `count` - Total number of scores
- `value` - Score value (for aggregations)

### scores-categorical
Query categorical score data. Same dimensions as scores-numeric except uses `stringValue` instead of `value`.

**Measures:**
- `count` - Total number of scores

## High Cardinality Dimensions
The following dimensions cannot be used as grouping dimensions in v2 metrics API as they can cause performance issues.
Use them in filters instead.

**observations view:**
- `id` - Use traceId filter to narrow down results
- `traceId` - Use traceId filter instead
- `userId` - Use userId filter instead
- `sessionId` - Use sessionId filter instead
- `parentObservationId` - Use parentObservationId filter instead

**scores-numeric / scores-categorical views:**
- `id` - Use specific filters to narrow down results
- `traceId` - Use traceId filter instead
- `userId` - Use userId filter instead
- `sessionId` - Use sessionId filter instead
- `observationId` - Use observationId filter instead

## Aggregations
Available aggregation functions: `sum`, `avg`, `count`, `max`, `min`, `p50`, `p75`, `p90`, `p95`, `p99`, `histogram`

## Time Granularities
Available granularities for timeDimension: `auto`, `minute`, `hour`, `day`, `week`, `month`
- `auto` bins the data into approximately 50 buckets based on the time range
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.MetricsV2MetricsRequest{
        Query: "query",
    }
client.Metricsv2.Metrics(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**query:** `string` 

JSON string containing the query parameters with the following structure:
```json
{
  "view": string,           // Required. One of "observations", "scores-numeric", "scores-categorical"
  "dimensions": [           // Optional. Default: []
    {
      "field": string       // Field to group by (see available dimensions above)
    }
  ],
  "metrics": [              // Required. At least one metric must be provided
    {
      "measure": string,    // What to measure (see available measures above)
      "aggregation": string // How to aggregate: "sum", "avg", "count", "max", "min", "p50", "p75", "p90", "p95", "p99", "histogram"
    }
  ],
  "filters": [              // Optional. Default: []
    {
      "column": string,     // Column to filter on (any dimension field)
      "operator": string,   // Operator based on type:
                            // - datetime: ">", "<", ">=", "<="
                            // - string: "=", "contains", "does not contain", "starts with", "ends with"
                            // - stringOptions: "any of", "none of"
                            // - arrayOptions: "any of", "none of", "all of"
                            // - number: "=", ">", "<", ">=", "<="
                            // - stringObject/numberObject: same as string/number with required "key"
                            // - boolean: "=", "<>"
                            // - null: "is null", "is not null"
      "value": any,         // Value to compare against
      "type": string,       // Data type: "datetime", "string", "number", "stringOptions", "categoryOptions", "arrayOptions", "stringObject", "numberObject", "boolean", "null"
      "key": string         // Required only for stringObject/numberObject types (e.g., metadata filtering)
    }
  ],
  "timeDimension": {        // Optional. Default: null. If provided, results will be grouped by time
    "granularity": string   // One of "auto", "minute", "hour", "day", "week", "month"
  },
  "fromTimestamp": string,  // Required. ISO datetime string for start of time range
  "toTimestamp": string,    // Required. ISO datetime string for end of time range (must be after fromTimestamp)
  "orderBy": [              // Optional. Default: null
    {
      "field": string,      // Field to order by (dimension or metric alias)
      "direction": string   // "asc" or "desc"
    }
  ],
  "config": {               // Optional. Query-specific configuration
    "bins": number,         // Optional. Number of bins for histogram aggregation (1-100), default: 10
    "row_limit": number     // Optional. Maximum number of rows to return (1-1000), default: 100
  }
}
```
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

## Metrics
<details><summary><code>client.Metrics.Metrics() -> *api.MetricsResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Get metrics from the Langfuse project using a query object.

Consider using the [v2 metrics endpoint](/api-reference#tag/metricsv2/GET/api/public/v2/metrics) for better performance.

For more details, see the [Metrics API documentation](https://langfuse.com/docs/metrics/features/metrics-api).
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.MetricsMetricsRequest{
        Query: "query",
    }
client.Metrics.Metrics(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**query:** `string` 

JSON string containing the query parameters with the following structure:
```json
{
  "view": string,           // Required. One of "traces", "observations", "scores-numeric", "scores-categorical"
  "dimensions": [           // Optional. Default: []
    {
      "field": string       // Field to group by, e.g. "name", "userId", "sessionId"
    }
  ],
  "metrics": [              // Required. At least one metric must be provided
    {
      "measure": string,    // What to measure, e.g. "count", "latency", "value"
      "aggregation": string // How to aggregate, e.g. "count", "sum", "avg", "p95", "histogram"
    }
  ],
  "filters": [              // Optional. Default: []
    {
      "column": string,     // Column to filter on
      "operator": string,   // Operator, e.g. "=", ">", "<", "contains"
      "value": any,         // Value to compare against
      "type": string,       // Data type, e.g. "string", "number", "stringObject"
      "key": string         // Required only when filtering on metadata
    }
  ],
  "timeDimension": {        // Optional. Default: null. If provided, results will be grouped by time
    "granularity": string   // One of "minute", "hour", "day", "week", "month", "auto"
  },
  "fromTimestamp": string,  // Required. ISO datetime string for start of time range
  "toTimestamp": string,    // Required. ISO datetime string for end of time range
  "orderBy": [              // Optional. Default: null
    {
      "field": string,      // Field to order by
      "direction": string   // "asc" or "desc"
    }
  ],
  "config": {               // Optional. Query-specific configuration
    "bins": number,         // Optional. Number of bins for histogram (1-100), default: 10
    "row_limit": number     // Optional. Row limit for results (1-1000)
  }
}
```
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

## Models
<details><summary><code>client.Models.List() -> *api.PaginatedModels</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Get all models
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.ModelsListRequest{}
client.Models.List(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**page:** `*int` — page number, starts at 1
    
</dd>
</dl>

<dl>
<dd>

**limit:** `*int` — limit of items per page
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Models.Create(request) -> *api.Model</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Create a model
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.CreateModelRequest{
        ModelName: "modelName",
        MatchPattern: "matchPattern",
    }
client.Models.Create(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**modelName:** `string` — Name of the model definition. If multiple with the same name exist, they are applied in the following order: (1) custom over built-in, (2) newest according to startTime where model.startTime<observation.startTime
    
</dd>
</dl>

<dl>
<dd>

**matchPattern:** `string` — Regex pattern which matches this model definition to generation.model. Useful in case of fine-tuned models. If you want to exact match, use `(?i)^modelname$`
    
</dd>
</dl>

<dl>
<dd>

**startDate:** `*time.Time` — Apply only to generations which are newer than this ISO date.
    
</dd>
</dl>

<dl>
<dd>

**unit:** `*api.ModelUsageUnit` — Unit used by this model.
    
</dd>
</dl>

<dl>
<dd>

**inputPrice:** `*float64` — Deprecated. Use 'pricingTiers' instead. Price (USD) per input unit. Creates a default tier if pricingTiers not provided.
    
</dd>
</dl>

<dl>
<dd>

**outputPrice:** `*float64` — Deprecated. Use 'pricingTiers' instead. Price (USD) per output unit. Creates a default tier if pricingTiers not provided.
    
</dd>
</dl>

<dl>
<dd>

**totalPrice:** `*float64` — Deprecated. Use 'pricingTiers' instead. Price (USD) per total units. Cannot be set if input or output price is set. Creates a default tier if pricingTiers not provided.
    
</dd>
</dl>

<dl>
<dd>

**pricingTiers:** `[]*api.PricingTierInput` 

Optional. Array of pricing tiers for this model.

Use pricing tiers for all models - both those with threshold-based pricing variations and those with simple flat pricing:

- For models with standard flat pricing: Create a single default tier with your prices
  (e.g., one tier with isDefault=true, priority=0, conditions=[], and your standard prices)

- For models with threshold-based pricing: Create a default tier plus additional conditional tiers
  (e.g., default tier for standard usage + high-volume tier for usage above certain thresholds)

Requirements:
- Cannot be provided with flat prices (inputPrice/outputPrice/totalPrice) - use one approach or the other
- Must include exactly one default tier with isDefault=true, priority=0, and conditions=[]
- All tier names and priorities must be unique within the model
- Each tier must define at least one price

If omitted, you must provide flat prices instead (inputPrice/outputPrice/totalPrice),
which will automatically create a single default tier named "Standard".
    
</dd>
</dl>

<dl>
<dd>

**tokenizerID:** `*string` — Optional. Tokenizer to be applied to observations which match to this model. See docs for more details.
    
</dd>
</dl>

<dl>
<dd>

**tokenizerConfig:** `any` — Optional. Configuration for the selected tokenizer. Needs to be JSON. See docs for more details.
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Models.Get(ID) -> *api.Model</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Get a model
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.ModelsGetRequest{
        ID: "id",
    }
client.Models.Get(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**id:** `string` 
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Models.Delete(ID) -> error</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Delete a model. Cannot delete models managed by Langfuse. You can create your own definition with the same modelName to override the definition though.
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.ModelsDeleteRequest{
        ID: "id",
    }
client.Models.Delete(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**id:** `string` 
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

## Observationsv2
<details><summary><code>client.Observationsv2.Getmany() -> *api.ObservationsV2Response</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Get a list of observations with cursor-based pagination and flexible field selection.

## Cursor-based Pagination
This endpoint uses cursor-based pagination for efficient traversal of large datasets.
The cursor is returned in the response metadata and should be passed in subsequent requests
to retrieve the next page of results.

## Field Selection
Use the `fields` parameter to control which observation fields are returned:
- `core` - Always included: id, traceId, startTime, endTime, projectId, parentObservationId, type
- `basic` - name, level, statusMessage, version, environment, bookmarked, public, userId, sessionId
- `time` - completionStartTime, createdAt, updatedAt
- `io` - input, output
- `metadata` - metadata (truncated to 200 chars by default, use `expandMetadata` to get full values)
- `model` - providedModelName, internalModelId, modelParameters
- `usage` - usageDetails, costDetails, totalCost
- `prompt` - promptId, promptName, promptVersion
- `metrics` - latency, timeToFirstToken

If not specified, `core` and `basic` field groups are returned.

## Filters
Multiple filtering options are available via query parameters or the structured `filter` parameter.
When using the `filter` parameter, it takes precedence over individual query parameter filters.
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.ObservationsV2GetManyRequest{}
client.Observationsv2.Getmany(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**fields:** `*string` 

Comma-separated list of field groups to include in the response.
Available groups: core, basic, time, io, metadata, model, usage, prompt, metrics.
If not specified, `core` and `basic` field groups are returned.
Example: "basic,usage,model"
    
</dd>
</dl>

<dl>
<dd>

**expandMetadata:** `*string` 

Comma-separated list of metadata keys to return non-truncated.
By default, metadata values over 200 characters are truncated.
Use this parameter to retrieve full values for specific keys.
Example: "key1,key2"
    
</dd>
</dl>

<dl>
<dd>

**limit:** `*int` — Number of items to return per page. Maximum 1000, default 50.
    
</dd>
</dl>

<dl>
<dd>

**cursor:** `*string` — Base64-encoded cursor for pagination. Use the cursor from the previous response to get the next page.
    
</dd>
</dl>

<dl>
<dd>

**parseIoAsJSON:** `*bool` 

Set to `true` to parse input/output fields as JSON, or `false` to return raw strings.
Defaults to `false` if not provided.
    
</dd>
</dl>

<dl>
<dd>

**name:** `*string` 
    
</dd>
</dl>

<dl>
<dd>

**userID:** `*string` 
    
</dd>
</dl>

<dl>
<dd>

**type_:** `*string` — Filter by observation type (e.g., "GENERATION", "SPAN", "EVENT", "AGENT", "TOOL", "CHAIN", "RETRIEVER", "EVALUATOR", "EMBEDDING", "GUARDRAIL")
    
</dd>
</dl>

<dl>
<dd>

**traceID:** `*string` 
    
</dd>
</dl>

<dl>
<dd>

**level:** `*api.ObservationLevel` — Optional filter for observations with a specific level (e.g. "DEBUG", "DEFAULT", "WARNING", "ERROR").
    
</dd>
</dl>

<dl>
<dd>

**parentObservationID:** `*string` 
    
</dd>
</dl>

<dl>
<dd>

**environment:** `*string` — Optional filter for observations where the environment is one of the provided values.
    
</dd>
</dl>

<dl>
<dd>

**fromStartTime:** `*time.Time` — Retrieve only observations with a start_time on or after this datetime (ISO 8601).
    
</dd>
</dl>

<dl>
<dd>

**toStartTime:** `*time.Time` — Retrieve only observations with a start_time before this datetime (ISO 8601).
    
</dd>
</dl>

<dl>
<dd>

**version:** `*string` — Optional filter to only include observations with a certain version.
    
</dd>
</dl>

<dl>
<dd>

**filter:** `*string` 

JSON string containing an array of filter conditions. When provided, this takes precedence over query parameter filters (userId, name, type, level, environment, fromStartTime, ...).

## Filter Structure
Each filter condition has the following structure:
```json
[
  {
    "type": string,           // Required. One of: "datetime", "string", "number", "stringOptions", "categoryOptions", "arrayOptions", "stringObject", "numberObject", "boolean", "null"
    "column": string,         // Required. Column to filter on (see available columns below)
    "operator": string,       // Required. Operator based on type:
                              // - datetime: ">", "<", ">=", "<="
                              // - string: "=", "contains", "does not contain", "starts with", "ends with"
                              // - stringOptions: "any of", "none of"
                              // - categoryOptions: "any of", "none of"
                              // - arrayOptions: "any of", "none of", "all of"
                              // - number: "=", ">", "<", ">=", "<="
                              // - stringObject: "=", "contains", "does not contain", "starts with", "ends with"
                              // - numberObject: "=", ">", "<", ">=", "<="
                              // - boolean: "=", "<>"
                              // - null: "is null", "is not null"
    "value": any,             // Required (except for null type). Value to compare against. Type depends on filter type
    "key": string             // Required only for stringObject, numberObject, and categoryOptions types when filtering on nested fields like metadata
  }
]
```

## Available Columns

### Core Observation Fields
- `id` (string) - Observation ID
- `type` (string) - Observation type (SPAN, GENERATION, EVENT)
- `name` (string) - Observation name
- `traceId` (string) - Associated trace ID
- `startTime` (datetime) - Observation start time
- `endTime` (datetime) - Observation end time
- `environment` (string) - Environment tag
- `level` (string) - Log level (DEBUG, DEFAULT, WARNING, ERROR)
- `statusMessage` (string) - Status message
- `version` (string) - Version tag
- `userId` (string) - User ID
- `sessionId` (string) - Session ID

### Trace-Related Fields
- `traceName` (string) - Name of the parent trace
- `traceTags` (arrayOptions) - Tags from the parent trace
- `tags` (arrayOptions) - Alias for traceTags

### Performance Metrics
- `latency` (number) - Latency in seconds (calculated: end_time - start_time)
- `timeToFirstToken` (number) - Time to first token in seconds
- `tokensPerSecond` (number) - Output tokens per second

### Token Usage
- `inputTokens` (number) - Number of input tokens
- `outputTokens` (number) - Number of output tokens
- `totalTokens` (number) - Total tokens (alias: `tokens`)

### Cost Metrics
- `inputCost` (number) - Input cost in USD
- `outputCost` (number) - Output cost in USD
- `totalCost` (number) - Total cost in USD

### Model Information
- `model` (string) - Provided model name (alias: `providedModelName`)
- `promptName` (string) - Associated prompt name
- `promptVersion` (number) - Associated prompt version

### Structured Data
- `metadata` (stringObject/numberObject/categoryOptions) - Metadata key-value pairs. Use `key` parameter to filter on specific metadata keys.

## Filter Examples
```json
[
  {
    "type": "string",
    "column": "type",
    "operator": "=",
    "value": "GENERATION"
  },
  {
    "type": "number",
    "column": "latency",
    "operator": ">=",
    "value": 2.5
  },
  {
    "type": "stringObject",
    "column": "metadata",
    "key": "environment",
    "operator": "=",
    "value": "production"
  }
]
```
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

## Observations
<details><summary><code>client.Observations.Get(ObservationID) -> *api.ObservationsView</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Get a observation
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.ObservationsGetRequest{
        ObservationID: "observationId",
    }
client.Observations.Get(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**observationID:** `string` — The unique langfuse identifier of an observation, can be an event, span or generation
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Observations.Getmany() -> *api.ObservationsViews</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Get a list of observations.

Consider using the [v2 observations endpoint](/api-reference#tag/observationsv2/GET/api/public/v2/observations) for cursor-based pagination and field selection.
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.ObservationsGetManyRequest{}
client.Observations.Getmany(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**page:** `*int` — Page number, starts at 1.
    
</dd>
</dl>

<dl>
<dd>

**limit:** `*int` — Limit of items per page. If you encounter api issues due to too large page sizes, try to reduce the limit.
    
</dd>
</dl>

<dl>
<dd>

**name:** `*string` 
    
</dd>
</dl>

<dl>
<dd>

**userID:** `*string` 
    
</dd>
</dl>

<dl>
<dd>

**type_:** `*string` 
    
</dd>
</dl>

<dl>
<dd>

**traceID:** `*string` 
    
</dd>
</dl>

<dl>
<dd>

**level:** `*api.ObservationLevel` — Optional filter for observations with a specific level (e.g. "DEBUG", "DEFAULT", "WARNING", "ERROR").
    
</dd>
</dl>

<dl>
<dd>

**parentObservationID:** `*string` 
    
</dd>
</dl>

<dl>
<dd>

**environment:** `*string` — Optional filter for observations where the environment is one of the provided values.
    
</dd>
</dl>

<dl>
<dd>

**fromStartTime:** `*time.Time` — Retrieve only observations with a start_time on or after this datetime (ISO 8601).
    
</dd>
</dl>

<dl>
<dd>

**toStartTime:** `*time.Time` — Retrieve only observations with a start_time before this datetime (ISO 8601).
    
</dd>
</dl>

<dl>
<dd>

**version:** `*string` — Optional filter to only include observations with a certain version.
    
</dd>
</dl>

<dl>
<dd>

**filter:** `*string` 

JSON string containing an array of filter conditions. When provided, this takes precedence over query parameter filters (userId, name, type, level, environment, fromStartTime, ...).

## Filter Structure
Each filter condition has the following structure:
```json
[
  {
    "type": string,           // Required. One of: "datetime", "string", "number", "stringOptions", "categoryOptions", "arrayOptions", "stringObject", "numberObject", "boolean", "null"
    "column": string,         // Required. Column to filter on (see available columns below)
    "operator": string,       // Required. Operator based on type:
                              // - datetime: ">", "<", ">=", "<="
                              // - string: "=", "contains", "does not contain", "starts with", "ends with"
                              // - stringOptions: "any of", "none of"
                              // - categoryOptions: "any of", "none of"
                              // - arrayOptions: "any of", "none of", "all of"
                              // - number: "=", ">", "<", ">=", "<="
                              // - stringObject: "=", "contains", "does not contain", "starts with", "ends with"
                              // - numberObject: "=", ">", "<", ">=", "<="
                              // - boolean: "=", "<>"
                              // - null: "is null", "is not null"
    "value": any,             // Required (except for null type). Value to compare against. Type depends on filter type
    "key": string             // Required only for stringObject, numberObject, and categoryOptions types when filtering on nested fields like metadata
  }
]
```

## Available Columns

### Core Observation Fields
- `id` (string) - Observation ID
- `type` (string) - Observation type (SPAN, GENERATION, EVENT)
- `name` (string) - Observation name
- `traceId` (string) - Associated trace ID
- `startTime` (datetime) - Observation start time
- `endTime` (datetime) - Observation end time
- `environment` (string) - Environment tag
- `level` (string) - Log level (DEBUG, DEFAULT, WARNING, ERROR)
- `statusMessage` (string) - Status message
- `version` (string) - Version tag

### Performance Metrics
- `latency` (number) - Latency in seconds (calculated: end_time - start_time)
- `timeToFirstToken` (number) - Time to first token in seconds
- `tokensPerSecond` (number) - Output tokens per second

### Token Usage
- `inputTokens` (number) - Number of input tokens
- `outputTokens` (number) - Number of output tokens
- `totalTokens` (number) - Total tokens (alias: `tokens`)

### Cost Metrics
- `inputCost` (number) - Input cost in USD
- `outputCost` (number) - Output cost in USD
- `totalCost` (number) - Total cost in USD

### Model Information
- `model` (string) - Provided model name
- `promptName` (string) - Associated prompt name
- `promptVersion` (number) - Associated prompt version

### Structured Data
- `metadata` (stringObject/numberObject/categoryOptions) - Metadata key-value pairs. Use `key` parameter to filter on specific metadata keys.

### Associated Trace Fields (requires join with traces table)
- `userId` (string) - User ID from associated trace
- `traceName` (string) - Name from associated trace
- `traceEnvironment` (string) - Environment from associated trace
- `traceTags` (arrayOptions) - Tags from associated trace

## Filter Examples
```json
[
  {
    "type": "string",
    "column": "type",
    "operator": "=",
    "value": "GENERATION"
  },
  {
    "type": "number",
    "column": "latency",
    "operator": ">=",
    "value": 2.5
  },
  {
    "type": "stringObject",
    "column": "metadata",
    "key": "environment",
    "operator": "=",
    "value": "production"
  }
]
```
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

## Opentelemetry
<details><summary><code>client.Opentelemetry.Exporttraces(request) -> *api.OtelTraceResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

**OpenTelemetry Traces Ingestion Endpoint**

This endpoint implements the OTLP/HTTP specification for trace ingestion, providing native OpenTelemetry integration for Langfuse Observability.

**Supported Formats:**
- Binary Protobuf: `Content-Type: application/x-protobuf`
- JSON Protobuf: `Content-Type: application/json`
- Supports gzip compression via `Content-Encoding: gzip` header

**Specification Compliance:**
- Conforms to [OTLP/HTTP Trace Export](https://opentelemetry.io/docs/specs/otlp/#otlphttp)
- Implements `ExportTraceServiceRequest` message format

**Documentation:**
- Integration guide: https://langfuse.com/integrations/native/opentelemetry
- Data model: https://langfuse.com/docs/observability/data-model
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.OpentelemetryExportTracesRequest{
        ResourceSpans: []*api.OtelResourceSpan{
            &api.OtelResourceSpan{
                Resource: &api.OtelResource{
                    Attributes: []*api.OtelAttribute{
                        &api.OtelAttribute{
                            Key: api.String(
                                "service.name",
                            ),
                            Value: &api.OtelAttributeValue{
                                StringValue: api.String(
                                    "my-service",
                                ),
                            },
                        },
                        &api.OtelAttribute{
                            Key: api.String(
                                "service.version",
                            ),
                            Value: &api.OtelAttributeValue{
                                StringValue: api.String(
                                    "1.0.0",
                                ),
                            },
                        },
                    },
                },
                ScopeSpans: []*api.OtelScopeSpan{
                    &api.OtelScopeSpan{
                        Scope: &api.OtelScope{
                            Name: api.String(
                                "langfuse-sdk",
                            ),
                            Version: api.String(
                                "2.60.3",
                            ),
                        },
                        Spans: []*api.OtelSpan{
                            &api.OtelSpan{
                                TraceID: "0123456789abcdef0123456789abcdef",
                                SpanID: "0123456789abcdef",
                                Name: api.String(
                                    "my-operation",
                                ),
                                Kind: api.Int(
                                    1,
                                ),
                                StartTimeUnixNano: "1747872000000000000",
                                EndTimeUnixNano: "1747872001000000000",
                                Attributes: []*api.OtelAttribute{
                                    &api.OtelAttribute{
                                        Key: api.String(
                                            "langfuse.observation.type",
                                        ),
                                        Value: &api.OtelAttributeValue{
                                            StringValue: api.String(
                                                "generation",
                                            ),
                                        },
                                    },
                                },
                                Status: map[string]any{},
                            },
                        },
                    },
                },
            },
        },
    }
client.Opentelemetry.Exporttraces(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**resourceSpans:** `[]*api.OtelResourceSpan` — Array of resource spans containing trace data as defined in the OTLP specification
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

## Organizations
<details><summary><code>client.Organizations.Getorganizationmemberships() -> *api.MembershipsResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Get all memberships for the organization associated with the API key (requires organization-scoped API key)
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
client.Organizations.Getorganizationmemberships(
        context.TODO(),
    )
}
```
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Organizations.Updateorganizationmembership(request) -> *api.MembershipResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Create or update a membership for the organization associated with the API key (requires organization-scoped API key)
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.MembershipRequest{
        UserID: "userId",
        Role: api.MembershipRoleOwner,
    }
client.Organizations.Updateorganizationmembership(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**request:** `*api.MembershipRequest` 
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Organizations.Deleteorganizationmembership(request) -> *api.MembershipDeletionResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Delete a membership from the organization associated with the API key (requires organization-scoped API key)
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.DeleteMembershipRequest{
        UserID: "userId",
    }
client.Organizations.Deleteorganizationmembership(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**request:** `*api.DeleteMembershipRequest` 
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Organizations.Getprojectmemberships(ProjectID) -> *api.MembershipsResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Get all memberships for a specific project (requires organization-scoped API key)
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.OrganizationsGetProjectMembershipsRequest{
        ProjectID: "projectId",
    }
client.Organizations.Getprojectmemberships(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**projectID:** `string` 
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Organizations.Updateprojectmembership(ProjectID, request) -> *api.MembershipResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Create or update a membership for a specific project (requires organization-scoped API key). The user must already be a member of the organization.
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.OrganizationsUpdateProjectMembershipRequest{
        ProjectID: "projectId",
        Body: &api.MembershipRequest{
            UserID: "userId",
            Role: api.MembershipRoleOwner,
        },
    }
client.Organizations.Updateprojectmembership(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**projectID:** `string` 
    
</dd>
</dl>

<dl>
<dd>

**request:** `*api.MembershipRequest` 
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Organizations.Deleteprojectmembership(ProjectID, request) -> *api.MembershipDeletionResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Delete a membership from a specific project (requires organization-scoped API key). The user must be a member of the organization.
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.OrganizationsDeleteProjectMembershipRequest{
        ProjectID: "projectId",
        Body: &api.DeleteMembershipRequest{
            UserID: "userId",
        },
    }
client.Organizations.Deleteprojectmembership(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**projectID:** `string` 
    
</dd>
</dl>

<dl>
<dd>

**request:** `*api.DeleteMembershipRequest` 
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Organizations.Getorganizationprojects() -> *api.OrganizationProjectsResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Get all projects for the organization associated with the API key (requires organization-scoped API key)
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
client.Organizations.Getorganizationprojects(
        context.TODO(),
    )
}
```
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Organizations.Getorganizationapikeys() -> *api.OrganizationAPIKeysResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Get all API keys for the organization associated with the API key (requires organization-scoped API key)
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
client.Organizations.Getorganizationapikeys(
        context.TODO(),
    )
}
```
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

## Projects
<details><summary><code>client.Projects.Get() -> *api.Projects</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Get Project associated with API key (requires project-scoped API key). You can use GET /api/public/organizations/projects to get all projects with an organization-scoped key.
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
client.Projects.Get(
        context.TODO(),
    )
}
```
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Projects.Create(request) -> *api.Project</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Create a new project (requires organization-scoped API key)
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.ProjectsCreateRequest{
        Name: "name",
        Retention: 1,
    }
client.Projects.Create(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**name:** `string` 
    
</dd>
</dl>

<dl>
<dd>

**metadata:** `map[string]any` — Optional metadata for the project
    
</dd>
</dl>

<dl>
<dd>

**retention:** `int` — Number of days to retain data. Must be 0 or at least 3 days. Requires data-retention entitlement for non-zero values. Optional.
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Projects.Update(ProjectID, request) -> *api.Project</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Update a project by ID (requires organization-scoped API key).
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.ProjectsUpdateRequest{
        ProjectID: "projectId",
        Name: "name",
    }
client.Projects.Update(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**projectID:** `string` 
    
</dd>
</dl>

<dl>
<dd>

**name:** `string` 
    
</dd>
</dl>

<dl>
<dd>

**metadata:** `map[string]any` — Optional metadata for the project
    
</dd>
</dl>

<dl>
<dd>

**retention:** `*int` 

Number of days to retain data.
Must be 0 or at least 3 days.
Requires data-retention entitlement for non-zero values.
Optional. Will retain existing retention setting if omitted.
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Projects.Delete(ProjectID) -> *api.ProjectDeletionResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Delete a project by ID (requires organization-scoped API key). Project deletion is processed asynchronously.
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.ProjectsDeleteRequest{
        ProjectID: "projectId",
    }
client.Projects.Delete(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**projectID:** `string` 
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Projects.Getapikeys(ProjectID) -> *api.APIKeyList</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Get all API keys for a project (requires organization-scoped API key)
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.ProjectsGetAPIKeysRequest{
        ProjectID: "projectId",
    }
client.Projects.Getapikeys(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**projectID:** `string` 
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Projects.Createapikey(ProjectID, request) -> *api.APIKeyResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Create a new API key for a project (requires organization-scoped API key)
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.ProjectsCreateAPIKeyRequest{
        ProjectID: "projectId",
    }
client.Projects.Createapikey(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**projectID:** `string` 
    
</dd>
</dl>

<dl>
<dd>

**note:** `*string` — Optional note for the API key
    
</dd>
</dl>

<dl>
<dd>

**publicKey:** `*string` — Optional predefined public key. Must start with 'pk-lf-'. If provided, secretKey must also be provided.
    
</dd>
</dl>

<dl>
<dd>

**secretKey:** `*string` — Optional predefined secret key. Must start with 'sk-lf-'. If provided, publicKey must also be provided.
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Projects.Deleteapikey(ProjectID, APIKeyID) -> *api.APIKeyDeletionResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Delete an API key for a project (requires organization-scoped API key)
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.ProjectsDeleteAPIKeyRequest{
        ProjectID: "projectId",
        APIKeyID: "apiKeyId",
    }
client.Projects.Deleteapikey(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**projectID:** `string` 
    
</dd>
</dl>

<dl>
<dd>

**apiKeyID:** `string` 
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

## Promptversion
<details><summary><code>client.Promptversion.Update(Name, Version, request) -> *api.Prompt</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Update labels for a specific prompt version
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.PromptVersionUpdateRequest{
        Name: "name",
        Version: 1,
        NewLabels: []string{
            "newLabels",
        },
    }
client.Promptversion.Update(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**name:** `string` 

The name of the prompt. If the prompt is in a folder (e.g., "folder/subfolder/prompt-name"), 
the folder path must be URL encoded.
    
</dd>
</dl>

<dl>
<dd>

**version:** `int` — Version of the prompt to update
    
</dd>
</dl>

<dl>
<dd>

**newLabels:** `[]string` — New labels for the prompt version. Labels are unique across versions. The "latest" label is reserved and managed by Langfuse.
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

## Prompts
<details><summary><code>client.Prompts.Get(PromptName) -> *api.Prompt</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Get a prompt
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.PromptsGetRequest{
        PromptName: "promptName",
    }
client.Prompts.Get(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**promptName:** `string` 

The name of the prompt. If the prompt is in a folder (e.g., "folder/subfolder/prompt-name"), 
the folder path must be URL encoded.
    
</dd>
</dl>

<dl>
<dd>

**version:** `*int` — Version of the prompt to be retrieved.
    
</dd>
</dl>

<dl>
<dd>

**label:** `*string` — Label of the prompt to be retrieved. Defaults to "production" if no label or version is set.
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Prompts.Delete(PromptName) -> error</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Delete prompt versions. If neither version nor label is specified, all versions of the prompt are deleted.
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.PromptsDeleteRequest{
        PromptName: "promptName",
    }
client.Prompts.Delete(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**promptName:** `string` — The name of the prompt
    
</dd>
</dl>

<dl>
<dd>

**label:** `*string` — Optional label to filter deletion. If specified, deletes all prompt versions that have this label.
    
</dd>
</dl>

<dl>
<dd>

**version:** `*int` — Optional version to filter deletion. If specified, deletes only this specific version of the prompt.
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Prompts.List() -> *api.PromptMetaListResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Get a list of prompt names with versions and labels
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.PromptsListRequest{}
client.Prompts.List(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**name:** `*string` 
    
</dd>
</dl>

<dl>
<dd>

**label:** `*string` 
    
</dd>
</dl>

<dl>
<dd>

**tag:** `*string` 
    
</dd>
</dl>

<dl>
<dd>

**page:** `*int` — page number, starts at 1
    
</dd>
</dl>

<dl>
<dd>

**limit:** `*int` — limit of items per page
    
</dd>
</dl>

<dl>
<dd>

**fromUpdatedAt:** `*time.Time` — Optional filter to only include prompt versions created/updated on or after a certain datetime (ISO 8601)
    
</dd>
</dl>

<dl>
<dd>

**toUpdatedAt:** `*time.Time` — Optional filter to only include prompt versions created/updated before a certain datetime (ISO 8601)
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Prompts.Create(request) -> *api.Prompt</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Create a new version for the prompt with the given `name`
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.CreatePromptRequest{
        CreatePromptRequestZero: &api.CreatePromptRequestZero{
            Name: "name",
            Prompt: []*api.ChatMessageWithPlaceholders{
                &api.ChatMessageWithPlaceholders{
                    ChatMessageWithPlaceholdersZero: &api.ChatMessageWithPlaceholdersZero{
                        Role: "role",
                        Content: "content",
                    },
                },
            },
        },
    }
client.Prompts.Create(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**request:** `*api.CreatePromptRequest` 
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

## SCIM
<details><summary><code>client.SCIM.Getserviceproviderconfig() -> *api.ServiceProviderConfig</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Get SCIM Service Provider Configuration (requires organization-scoped API key)
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
client.SCIM.Getserviceproviderconfig(
        context.TODO(),
    )
}
```
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.SCIM.Getresourcetypes() -> *api.ResourceTypesResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Get SCIM Resource Types (requires organization-scoped API key)
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
client.SCIM.Getresourcetypes(
        context.TODO(),
    )
}
```
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.SCIM.Getschemas() -> *api.SchemasResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Get SCIM Schemas (requires organization-scoped API key)
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
client.SCIM.Getschemas(
        context.TODO(),
    )
}
```
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.SCIM.Listusers() -> *api.SCIMUsersListResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

List users in the organization (requires organization-scoped API key)
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.SCIMListUsersRequest{}
client.SCIM.Listusers(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**filter:** `*string` — Filter expression (e.g. userName eq "value")
    
</dd>
</dl>

<dl>
<dd>

**startIndex:** `*int` — 1-based index of the first result to return (default 1)
    
</dd>
</dl>

<dl>
<dd>

**count:** `*int` — Maximum number of results to return (default 100)
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.SCIM.Createuser(request) -> *api.SCIMUser</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Create a new user in the organization (requires organization-scoped API key)
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.SCIMCreateUserRequest{
        UserName: "userName",
        Name: &api.SCIMName{},
    }
client.SCIM.Createuser(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**userName:** `string` — User's email address (required)
    
</dd>
</dl>

<dl>
<dd>

**name:** `*api.SCIMName` — User's name information
    
</dd>
</dl>

<dl>
<dd>

**emails:** `[]*api.SCIMEmail` — User's email addresses
    
</dd>
</dl>

<dl>
<dd>

**active:** `*bool` — Whether the user is active
    
</dd>
</dl>

<dl>
<dd>

**password:** `*string` — Initial password for the user
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.SCIM.Getuser(UserID) -> *api.SCIMUser</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Get a specific user by ID (requires organization-scoped API key)
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.SCIMGetUserRequest{
        UserID: "userId",
    }
client.SCIM.Getuser(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**userID:** `string` 
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.SCIM.Deleteuser(UserID) -> *api.EmptyResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Remove a user from the organization (requires organization-scoped API key). Note that this only removes the user from the organization but does not delete the user entity itself.
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.SCIMDeleteUserRequest{
        UserID: "userId",
    }
client.SCIM.Deleteuser(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**userID:** `string` 
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

## Scoreconfigs
<details><summary><code>client.Scoreconfigs.Get() -> *api.ScoreConfigs</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Get all score configs
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.ScoreConfigsGetRequest{}
client.Scoreconfigs.Get(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**page:** `*int` — Page number, starts at 1.
    
</dd>
</dl>

<dl>
<dd>

**limit:** `*int` — Limit of items per page. If you encounter api issues due to too large page sizes, try to reduce the limit
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Scoreconfigs.Create(request) -> *api.ScoreConfig</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Create a score configuration (config). Score configs are used to define the structure of scores
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.CreateScoreConfigRequest{
        Name: "name",
        DataType: api.ScoreConfigDataTypeNumeric,
    }
client.Scoreconfigs.Create(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**name:** `string` 
    
</dd>
</dl>

<dl>
<dd>

**dataType:** `*api.ScoreConfigDataType` 
    
</dd>
</dl>

<dl>
<dd>

**categories:** `[]*api.ConfigCategory` — Configure custom categories for categorical scores. Pass a list of objects with `label` and `value` properties. Categories are autogenerated for boolean configs and cannot be passed
    
</dd>
</dl>

<dl>
<dd>

**minValue:** `*float64` — Configure a minimum value for numerical scores. If not set, the minimum value defaults to -∞
    
</dd>
</dl>

<dl>
<dd>

**maxValue:** `*float64` — Configure a maximum value for numerical scores. If not set, the maximum value defaults to +∞
    
</dd>
</dl>

<dl>
<dd>

**description:** `*string` — Description is shown across the Langfuse UI and can be used to e.g. explain the config categories in detail, why a numeric range was set, or provide additional context on config name or usage
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Scoreconfigs.GetByID(ConfigID) -> *api.ScoreConfig</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Get a score config
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.ScoreConfigsGetByIDRequest{
        ConfigID: "configId",
    }
client.Scoreconfigs.GetByID(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**configID:** `string` — The unique langfuse identifier of a score config
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Scoreconfigs.Update(ConfigID, request) -> *api.ScoreConfig</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Update a score config
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.UpdateScoreConfigRequest{
        ConfigID: "configId",
    }
client.Scoreconfigs.Update(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**configID:** `string` — The unique langfuse identifier of a score config
    
</dd>
</dl>

<dl>
<dd>

**isArchived:** `*bool` — The status of the score config showing if it is archived or not
    
</dd>
</dl>

<dl>
<dd>

**name:** `*string` — The name of the score config
    
</dd>
</dl>

<dl>
<dd>

**categories:** `[]*api.ConfigCategory` — Configure custom categories for categorical scores. Pass a list of objects with `label` and `value` properties. Categories are autogenerated for boolean configs and cannot be passed
    
</dd>
</dl>

<dl>
<dd>

**minValue:** `*float64` — Configure a minimum value for numerical scores. If not set, the minimum value defaults to -∞
    
</dd>
</dl>

<dl>
<dd>

**maxValue:** `*float64` — Configure a maximum value for numerical scores. If not set, the maximum value defaults to +∞
    
</dd>
</dl>

<dl>
<dd>

**description:** `*string` — Description is shown across the Langfuse UI and can be used to e.g. explain the config categories in detail, why a numeric range was set, or provide additional context on config name or usage
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

## Scorev2
<details><summary><code>client.Scorev2.Get() -> *api.GetScoresResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Get a list of scores (supports both trace and session scores)
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.ScoreV2GetRequest{}
client.Scorev2.Get(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**page:** `*int` — Page number, starts at 1.
    
</dd>
</dl>

<dl>
<dd>

**limit:** `*int` — Limit of items per page. If you encounter api issues due to too large page sizes, try to reduce the limit.
    
</dd>
</dl>

<dl>
<dd>

**userID:** `*string` — Retrieve only scores with this userId associated to the trace.
    
</dd>
</dl>

<dl>
<dd>

**name:** `*string` — Retrieve only scores with this name.
    
</dd>
</dl>

<dl>
<dd>

**fromTimestamp:** `*time.Time` — Optional filter to only include scores created on or after a certain datetime (ISO 8601)
    
</dd>
</dl>

<dl>
<dd>

**toTimestamp:** `*time.Time` — Optional filter to only include scores created before a certain datetime (ISO 8601)
    
</dd>
</dl>

<dl>
<dd>

**environment:** `*string` — Optional filter for scores where the environment is one of the provided values.
    
</dd>
</dl>

<dl>
<dd>

**source:** `*api.ScoreSource` — Retrieve only scores from a specific source.
    
</dd>
</dl>

<dl>
<dd>

**operator:** `*string` — Retrieve only scores with <operator> value.
    
</dd>
</dl>

<dl>
<dd>

**value:** `*float64` — Retrieve only scores with <operator> value.
    
</dd>
</dl>

<dl>
<dd>

**scoreIDs:** `*string` — Comma-separated list of score IDs to limit the results to.
    
</dd>
</dl>

<dl>
<dd>

**configID:** `*string` — Retrieve only scores with a specific configId.
    
</dd>
</dl>

<dl>
<dd>

**sessionID:** `*string` — Retrieve only scores with a specific sessionId.
    
</dd>
</dl>

<dl>
<dd>

**datasetRunID:** `*string` — Retrieve only scores with a specific datasetRunId.
    
</dd>
</dl>

<dl>
<dd>

**traceID:** `*string` — Retrieve only scores with a specific traceId.
    
</dd>
</dl>

<dl>
<dd>

**queueID:** `*string` — Retrieve only scores with a specific annotation queueId.
    
</dd>
</dl>

<dl>
<dd>

**dataType:** `*api.ScoreDataType` — Retrieve only scores with a specific dataType.
    
</dd>
</dl>

<dl>
<dd>

**traceTags:** `*string` — Only scores linked to traces that include all of these tags will be returned.
    
</dd>
</dl>

<dl>
<dd>

**fields:** `*string` — Comma-separated list of field groups to include in the response. Available field groups: 'score' (core score fields), 'trace' (trace properties: userId, tags, environment). If not specified, both 'score' and 'trace' are returned by default. Example: 'score' to exclude trace data, 'score,trace' to include both. Note: When filtering by trace properties (using userId or traceTags parameters), the 'trace' field group must be included, otherwise a 400 error will be returned.
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Scorev2.GetByID(ScoreID) -> *api.Score</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Get a score (supports both trace and session scores)
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.ScoreV2GetByIDRequest{
        ScoreID: "scoreId",
    }
client.Scorev2.GetByID(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**scoreID:** `string` — The unique langfuse identifier of a score
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

## Score
<details><summary><code>client.Score.Create(request) -> *api.CreateScoreResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Create a score (supports both trace and session scores)
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.CreateScoreRequest{
        Name: "novelty",
        Value: &api.CreateScoreValue{
            Double: 1.1,
        },
    }
client.Score.Create(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**id:** `*string` 
    
</dd>
</dl>

<dl>
<dd>

**traceID:** `*string` 
    
</dd>
</dl>

<dl>
<dd>

**sessionID:** `*string` 
    
</dd>
</dl>

<dl>
<dd>

**observationID:** `*string` 
    
</dd>
</dl>

<dl>
<dd>

**datasetRunID:** `*string` 
    
</dd>
</dl>

<dl>
<dd>

**name:** `string` 
    
</dd>
</dl>

<dl>
<dd>

**value:** `*api.CreateScoreValue` — The value of the score. Must be passed as string for categorical scores, and numeric for boolean and numeric scores. Boolean score values must equal either 1 or 0 (true or false)
    
</dd>
</dl>

<dl>
<dd>

**comment:** `*string` 
    
</dd>
</dl>

<dl>
<dd>

**metadata:** `map[string]any` 
    
</dd>
</dl>

<dl>
<dd>

**environment:** `*string` — The environment of the score. Can be any lowercase alphanumeric string with hyphens and underscores that does not start with 'langfuse'.
    
</dd>
</dl>

<dl>
<dd>

**queueID:** `*string` — The annotation queue referenced by the score. Indicates if score was initially created while processing annotation queue.
    
</dd>
</dl>

<dl>
<dd>

**dataType:** `*api.ScoreDataType` — The data type of the score. When passing a configId this field is inferred. Otherwise, this field must be passed or will default to numeric.
    
</dd>
</dl>

<dl>
<dd>

**configID:** `*string` — Reference a score config on a score. The unique langfuse identifier of a score config. When passing this field, the dataType and stringValue fields are automatically populated.
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Score.Delete(ScoreID) -> error</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Delete a score (supports both trace and session scores)
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.ScoreDeleteRequest{
        ScoreID: "scoreId",
    }
client.Score.Delete(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**scoreID:** `string` — The unique langfuse identifier of a score
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

## Sessions
<details><summary><code>client.Sessions.List() -> *api.PaginatedSessions</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Get sessions
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.SessionsListRequest{}
client.Sessions.List(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**page:** `*int` — Page number, starts at 1
    
</dd>
</dl>

<dl>
<dd>

**limit:** `*int` — Limit of items per page. If you encounter api issues due to too large page sizes, try to reduce the limit.
    
</dd>
</dl>

<dl>
<dd>

**fromTimestamp:** `*time.Time` — Optional filter to only include sessions created on or after a certain datetime (ISO 8601)
    
</dd>
</dl>

<dl>
<dd>

**toTimestamp:** `*time.Time` — Optional filter to only include sessions created before a certain datetime (ISO 8601)
    
</dd>
</dl>

<dl>
<dd>

**environment:** `*string` — Optional filter for sessions where the environment is one of the provided values.
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Sessions.Get(SessionID) -> *api.SessionWithTraces</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Get a session. Please note that `traces` on this endpoint are not paginated, if you plan to fetch large sessions, consider `GET /api/public/traces?sessionId=<sessionId>`
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.SessionsGetRequest{
        SessionID: "sessionId",
    }
client.Sessions.Get(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**sessionID:** `string` — The unique id of a session
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

## Trace
<details><summary><code>client.Trace.Get(TraceID) -> *api.TraceWithFullDetails</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Get a specific trace
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.TraceGetRequest{
        TraceID: "traceId",
    }
client.Trace.Get(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**traceID:** `string` — The unique langfuse identifier of a trace
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Trace.Delete(TraceID) -> *api.DeleteTraceResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Delete a specific trace
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.TraceDeleteRequest{
        TraceID: "traceId",
    }
client.Trace.Delete(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**traceID:** `string` — The unique langfuse identifier of the trace to delete
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Trace.List() -> *api.Traces</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Get list of traces
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.TraceListRequest{}
client.Trace.List(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**page:** `*int` — Page number, starts at 1
    
</dd>
</dl>

<dl>
<dd>

**limit:** `*int` — Limit of items per page. If you encounter api issues due to too large page sizes, try to reduce the limit.
    
</dd>
</dl>

<dl>
<dd>

**userID:** `*string` 
    
</dd>
</dl>

<dl>
<dd>

**name:** `*string` 
    
</dd>
</dl>

<dl>
<dd>

**sessionID:** `*string` 
    
</dd>
</dl>

<dl>
<dd>

**fromTimestamp:** `*time.Time` — Optional filter to only include traces with a trace.timestamp on or after a certain datetime (ISO 8601)
    
</dd>
</dl>

<dl>
<dd>

**toTimestamp:** `*time.Time` — Optional filter to only include traces with a trace.timestamp before a certain datetime (ISO 8601)
    
</dd>
</dl>

<dl>
<dd>

**orderBy:** `*string` — Format of the string [field].[asc/desc]. Fields: id, timestamp, name, userId, release, version, public, bookmarked, sessionId. Example: timestamp.asc
    
</dd>
</dl>

<dl>
<dd>

**tags:** `*string` — Only traces that include all of these tags will be returned.
    
</dd>
</dl>

<dl>
<dd>

**version:** `*string` — Optional filter to only include traces with a certain version.
    
</dd>
</dl>

<dl>
<dd>

**release:** `*string` — Optional filter to only include traces with a certain release.
    
</dd>
</dl>

<dl>
<dd>

**environment:** `*string` — Optional filter for traces where the environment is one of the provided values.
    
</dd>
</dl>

<dl>
<dd>

**fields:** `*string` — Comma-separated list of fields to include in the response. Available field groups: 'core' (always included), 'io' (input, output, metadata), 'scores', 'observations', 'metrics'. If not specified, all fields are returned. Example: 'core,scores,metrics'. Note: Excluded 'observations' or 'scores' fields return empty arrays; excluded 'metrics' returns -1 for 'totalCost' and 'latency'.
    
</dd>
</dl>

<dl>
<dd>

**filter:** `*string` 

JSON string containing an array of filter conditions. When provided, this takes precedence over query parameter filters (userId, name, sessionId, tags, version, release, environment, fromTimestamp, toTimestamp).

## Filter Structure
Each filter condition has the following structure:
```json
[
  {
    "type": string,           // Required. One of: "datetime", "string", "number", "stringOptions", "categoryOptions", "arrayOptions", "stringObject", "numberObject", "boolean", "null"
    "column": string,         // Required. Column to filter on (see available columns below)
    "operator": string,       // Required. Operator based on type:
                              // - datetime: ">", "<", ">=", "<="
                              // - string: "=", "contains", "does not contain", "starts with", "ends with"
                              // - stringOptions: "any of", "none of"
                              // - categoryOptions: "any of", "none of"
                              // - arrayOptions: "any of", "none of", "all of"
                              // - number: "=", ">", "<", ">=", "<="
                              // - stringObject: "=", "contains", "does not contain", "starts with", "ends with"
                              // - numberObject: "=", ">", "<", ">=", "<="
                              // - boolean: "=", "<>"
                              // - null: "is null", "is not null"
    "value": any,             // Required (except for null type). Value to compare against. Type depends on filter type
    "key": string             // Required only for stringObject, numberObject, and categoryOptions types when filtering on nested fields like metadata
  }
]
```

## Available Columns

### Core Trace Fields
- `id` (string) - Trace ID
- `name` (string) - Trace name
- `timestamp` (datetime) - Trace timestamp
- `userId` (string) - User ID
- `sessionId` (string) - Session ID
- `environment` (string) - Environment tag
- `version` (string) - Version tag
- `release` (string) - Release tag
- `tags` (arrayOptions) - Array of tags
- `bookmarked` (boolean) - Bookmark status

### Structured Data
- `metadata` (stringObject/numberObject/categoryOptions) - Metadata key-value pairs. Use `key` parameter to filter on specific metadata keys.

### Aggregated Metrics (from observations)
These metrics are aggregated from all observations within the trace:
- `latency` (number) - Latency in seconds (time from first observation start to last observation end)
- `inputTokens` (number) - Total input tokens across all observations
- `outputTokens` (number) - Total output tokens across all observations
- `totalTokens` (number) - Total tokens (alias: `tokens`)
- `inputCost` (number) - Total input cost in USD
- `outputCost` (number) - Total output cost in USD
- `totalCost` (number) - Total cost in USD

### Observation Level Aggregations
These fields aggregate observation levels within the trace:
- `level` (string) - Highest severity level (ERROR > WARNING > DEFAULT > DEBUG)
- `warningCount` (number) - Count of WARNING level observations
- `errorCount` (number) - Count of ERROR level observations
- `defaultCount` (number) - Count of DEFAULT level observations
- `debugCount` (number) - Count of DEBUG level observations

### Scores (requires join with scores table)
- `scores_avg` (number) - Average of numeric scores (alias: `scores`)
- `score_categories` (categoryOptions) - Categorical score values

## Filter Examples
```json
[
  {
    "type": "datetime",
    "column": "timestamp",
    "operator": ">=",
    "value": "2024-01-01T00:00:00Z"
  },
  {
    "type": "string",
    "column": "userId",
    "operator": "=",
    "value": "user-123"
  },
  {
    "type": "number",
    "column": "totalCost",
    "operator": ">=",
    "value": 0.01
  },
  {
    "type": "arrayOptions",
    "column": "tags",
    "operator": "all of",
    "value": ["production", "critical"]
  },
  {
    "type": "stringObject",
    "column": "metadata",
    "key": "customer_tier",
    "operator": "=",
    "value": "enterprise"
  }
]
```

## Performance Notes
- Filtering on `userId`, `sessionId`, or `metadata` may enable skip indexes for better query performance
- Score filters require a join with the scores table and may impact query performance
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>

<details><summary><code>client.Trace.Deletemultiple(request) -> *api.DeleteTraceResponse</code></summary>
<dl>
<dd>

#### 📝 Description

<dl>
<dd>

<dl>
<dd>

Delete multiple traces
</dd>
</dl>
</dd>
</dl>

#### 🔌 Usage

<dl>
<dd>

<dl>
<dd>

```go
request := &api.TraceDeleteMultipleRequest{
        TraceIDs: []string{
            "traceIds",
        },
    }
client.Trace.Deletemultiple(
        context.TODO(),
        request,
    )
}
```
</dd>
</dl>
</dd>
</dl>

#### ⚙️ Parameters

<dl>
<dd>

<dl>
<dd>

**traceIDs:** `[]string` — List of trace IDs to delete
    
</dd>
</dl>
</dd>
</dl>


</dd>
</dl>
</details>
