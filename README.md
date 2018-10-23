#### PDF文件数据脱敏处理(For learning)
> 用途：

`pdf文件内容的读取、缓存、敏感数据识别定位、脱敏处理后生成新pdf文件.`
 

> 实例： 

-  单页文件脱敏处理(手机号/身份证号)
![单页文件处理](https://github.com/gotoolkits/pdfDtProtector/blob/master/gif/pdf_single.gif)
-  多页文件脱敏处理 (手机号/身份证号)
![多页文件处理](https://github.com/gotoolkits/pdfDtProtector/blob/master/gif/pdf_multi.gif)

> 依赖：

- imagemagick 
- ghostscript

> 配置文件config.json说明：
```
{
  "settings": {            
    "imagePPI": 227,                   //PPI值
    "compressionQuality": 80,          //图像压缩比
    "maskRows": 50,                    //掩盖图高度pix
    "offset": 10                       //掩盖图Y偏移     
  },
  "rules": {
    "regxRule": [                      //匹配规则定义
        "\\D(13[0-9]|14[579]|15[0-3,5-9]|16[6]|17[0135678]|18[0-9]|19[89])\\d{8}",                     //手机号正则
        "\\D[1-9]\\d{5}[1-9]\\d{3}((0\\d)|(1[0-2]))(([0|1|2]\\d)|3[0-1])\\d{3}([0-9]|X)"  //18位身份证正则
    ]
  }
}
```

> TODO:

- 优化处理速度
- 支持中文匹配


