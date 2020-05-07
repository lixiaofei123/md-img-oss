package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

func uploadFileToAliOSS(bucket *oss.Bucket, filepath, domain, ossDir string, dirLevel int) (string, error) {
	file, err := os.Open(filepath)
	if err == nil {
		defer file.Close()
		arr := strings.Split(filepath, "/")
		objectName := arr[len(arr)-1]
		for i := 0; i < dirLevel; i++ {
			objectName = arr[len(arr)-i-2] + "/" + objectName
		}
		if ossDir != "" {
			objectName = ossDir + "/" + objectName
		}
		storageType := oss.ObjectStorageClass(oss.StorageStandard)
		objectACL := oss.ObjectACL(oss.ACLPublicRead)
		err = bucket.PutObject(objectName, file, storageType, objectACL)
		if err == nil {
			newpath := fmt.Sprintf("%v/%v", domain, objectName)
			return newpath, nil
		} else {
			return "", err
		}
	} else {
		return "", err
	}
}

func main() {
	var mddir string
	var endpoint string
	var accesskeyID string
	var accessKeySecret string
	var bucketName string
	var ossDir string
	var dirLevel int
	var domain string

	flag.Usage = func() {
		fmt.Println(`
这个工具的作用是扫描指定目录下的md文件，如果发现md中包含的图片为本地图片路径，则会自动将该图片上传至阿里云的OSS上，并将md中的图片的本地路径替换为阿里云OSS的地址
例如:
md-img-oss -mddir /home/mds -endpoint oss-cn-shenzhen.aliyuncs.com -accesskeyId xxxxxxx  -accessKeySecret xxxxxxx -bucketName xxxxxx
		`)
		flag.PrintDefaults()
	}
	flag.StringVar(&mddir, "mddir", "", "markdown文件所在的目录")
	flag.StringVar(&endpoint, "endpoint", "oss-cn-shenzhen.aliyuncs.com", "阿里云OSS的endpoint")
	flag.StringVar(&accesskeyID, "accesskeyId", "", "accessKeyId")
	flag.StringVar(&accessKeySecret, "accessKeySecret", "", "accessKeySecret")
	flag.StringVar(&bucketName, "bucketName", "", "bucket名称")
	flag.StringVar(&ossDir, "ossDir", "", "将数据上传到oss的此目录下")
	flag.StringVar(&domain, "domain", "", "OSS绑定的域名，此项为空则设置为阿里云oss的默认地址")
	flag.IntVar(&dirLevel, "dirLevel", 1, "保留本地路径的层级")
	flag.Parse()

	required := []string{"mddir", "endpoint", "accesskeyId", "accessKeySecret", "bucketName"}

	seen := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { seen[f.Name] = true })
	for _, req := range required {
		if !seen[req] {
			fmt.Fprintf(os.Stderr, "缺少%s参数", req)
			os.Exit(-1)
		}
	}

	// 去除ossDir前后的 /
	if strings.HasSuffix(ossDir, "/") {
		ossDir = ossDir[:len(ossDir)-1]
	}
	if strings.HasPrefix(ossDir, "/") {
		ossDir = ossDir[1:]
	}

	// 使用默认的domain
	if domain == "" {
		domain = fmt.Sprintf("https://%v.%v", bucketName, endpoint)
	}

	ossClient, err := oss.New(endpoint, accesskeyID, accessKeySecret)
	if err != nil {
		log.Fatalf("创建Client失败,原因: %v \n", err)
	}

	bucket, err := ossClient.Bucket(bucketName)
	if err != nil {
		log.Fatalf("访问Bucket失败,原因: %v \n", err)
	}

	re := regexp.MustCompile("\\!\\[(.*?)\\]\\((.*?)\\)")
	fileinfos, err := ioutil.ReadDir(mddir)
	if err == nil {
		for _, fileinfo := range fileinfos {
			if !fileinfo.IsDir() && strings.HasSuffix(fileinfo.Name(), ".md") {
				fmt.Println("扫描md文件:", fileinfo.Name())
				mdPath := mddir + "/" + fileinfo.Name()
				bytes, err := ioutil.ReadFile(mdPath)
				if err == nil {
					isReplace := false
					mdContent := string(bytes)
					matches := re.FindAllStringSubmatch(mdContent, -1)
					for _, match := range matches {
						title := match[1]
						sourcepath := match[2]
						if !strings.HasPrefix(sourcepath, "http") {
							path := sourcepath
							if !filepath.IsAbs(path) {
								path = mddir + "/" + path
							}
							fmt.Println("发现本地图片路径:", path)
							fmt.Println("将其上传至OSS")

							newpath, err := uploadFileToAliOSS(bucket, path, domain, ossDir, dirLevel)
							if err == nil {
								fmt.Println("上传成功，新的路径为", newpath)
								newre := regexp.MustCompile("\\!\\[" + title + "\\]\\(" + sourcepath + "\\)")
								mdContent = newre.ReplaceAllString(mdContent, fmt.Sprintf("![%v](%v)", title, newpath))
								isReplace = true
							} else {
								log.Fatalf("上传资源至OSS失败,原因: %v \n", err)
							}
						}
					}

					if isReplace {
						ioutil.WriteFile(mdPath, []byte(mdContent), 0755)
						fmt.Println("替换成功")
					} else {
						fmt.Println("无需替换任何图片")
					}
				}
			}
		}
	}
}
