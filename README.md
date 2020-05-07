这个工具的作用是扫描指定目录下的md文件，如果发现md中包含的图片为本地图片路径，则会自动将该图片上传至阿里云的OSS上，并将md中的图片的本地路径替换为阿里云OSS的地址
例如:
```bash
md-img-oss -mddir /home/mds -endpoint oss-cn-shenzhen.aliyuncs.com -accesskeyId xxxxxxx  -accessKeySecret xxxxxxx -bucketName xxxxxx
```

获取
```bash
go get github.com/lixiaofei123/md-img-oss
```

参数详细解释

```
-accessKeySecret string
    aliyun账号的accessKeySecret
-accesskeyId string
     aliyun账号的accessKeyId
-bucketName string
    bucket名称
-dirLevel int
    保留本地路径的层级 (default 1)
    例如，本地文件路径为 /a/b/c/d/e.jpg，如果dirLevel为1，则上传时指定的Object key为d/e.jpg，如果设置为2,则指定的Object key为c/d/e.jpg
-domain string
    OSS绑定的域名，此项为空则设置为阿里云oss的默认提供地址
-endpoint string
    阿里云OSS的endpoint (default "oss-cn-shenzhen.aliyuncs.com")
-mddir string
    markdown文件所在的目录
-ossDir string
    将数据上传到oss的此目录下
```

