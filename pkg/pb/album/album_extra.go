// Дополнительные типы для UploadPhoto RPC, не сгенерированные protoc.
// При регенерации proto — удалить этот файл и добавить сообщения в proto/album.proto.
package album

// UploadPhotoRequest — запрос на регистрацию уже сохранённого файла и связь с альбомом.
type UploadPhotoRequest struct {
	AlbumId  uint64 `protobuf:"varint,1,opt,name=album_id,json=albumId,proto3" json:"album_id,omitempty"`
	FilePath string `protobuf:"bytes,2,opt,name=file_path,json=filePath,proto3" json:"file_path,omitempty"`
}

func (x *UploadPhotoRequest) Reset()         {}
func (x *UploadPhotoRequest) String() string  { return x.FilePath }
func (x *UploadPhotoRequest) ProtoMessage()   {}

func (x *UploadPhotoRequest) GetAlbumId() uint64  { return x.AlbumId }
func (x *UploadPhotoRequest) GetFilePath() string  { return x.FilePath }

// UploadPhotoResponse — ответ: id созданной записи photo и url.
type UploadPhotoResponse struct {
	PhotoId uint64 `protobuf:"varint,1,opt,name=photo_id,json=photoId,proto3" json:"photo_id,omitempty"`
	FileUrl string `protobuf:"bytes,2,opt,name=file_url,json=fileUrl,proto3" json:"file_url,omitempty"`
}

func (x *UploadPhotoResponse) Reset()         {}
func (x *UploadPhotoResponse) String() string  { return x.FileUrl }
func (x *UploadPhotoResponse) ProtoMessage()   {}

func (x *UploadPhotoResponse) GetPhotoId() uint64  { return x.PhotoId }
func (x *UploadPhotoResponse) GetFileUrl() string   { return x.FileUrl }
