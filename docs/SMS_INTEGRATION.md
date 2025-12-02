# 短信验证码服务商接入指南

## 当前实现状态

**⚠️ 重要提示**：当前项目的短信验证码功能是**模拟实现**，仅用于开发和测试环境。生产环境必须接入真实的短信服务商。

### 当前实现方式

- ✅ 验证码生成、存储、验证逻辑已完整实现
- ✅ 防刷机制（1分钟内只能发送一次）
- ✅ Redis 缓存加速验证
- ✅ 验证码有效期管理（5分钟）
- ❌ **缺少真实的短信发送功能**（仅在日志中输出验证码）

## 为什么需要接入服务商？

1. **合规要求**：真实短信发送需要经过运营商和服务商
2. **用户体验**：用户需要收到真实的短信验证码
3. **安全性**：防止验证码泄露（日志输出不安全）
4. **生产可用**：模拟实现无法在生产环境使用

## 推荐的短信服务商

### 国内服务商

1. **阿里云短信服务**
   - 官方文档：https://help.aliyun.com/product/44282.html
   - 价格：约 0.045 元/条
   - 优点：稳定可靠，文档完善
   - SDK：`github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi`

2. **腾讯云短信服务**
   - 官方文档：https://cloud.tencent.com/document/product/382
   - 价格：约 0.045 元/条
   - 优点：与腾讯生态集成好
   - SDK：`github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms`

3. **七牛云短信服务**
   - 官方文档：https://developer.qiniu.com/sms
   - 价格：约 0.04 元/条
   - 优点：价格相对便宜

4. **容联云通讯**
   - 官方文档：https://doc.yuntongxun.com/
   - 价格：约 0.05 元/条
   - 优点：功能丰富

### 国际服务商

1. **Twilio**
   - 官方文档：https://www.twilio.com/docs/sms
   - 优点：全球覆盖，API 简单
   - SDK：`github.com/twilio/twilio-go`

2. **AWS SNS**
   - 官方文档：https://docs.aws.amazon.com/sns/
   - 优点：与 AWS 生态集成好

## 接入示例（阿里云短信）

### 1. 安装 SDK

```bash
go get github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi
```

### 2. 添加配置

在 `internal/config/config.go` 中添加短信配置：

```go
type Config struct {
    // ... 其他配置
    SMS SMSConfig
}

type SMSConfig struct {
    Provider    string // "aliyun", "tencent", "twilio" 等
    AccessKeyID string
    AccessKeySecret string
    SignName    string // 短信签名
    TemplateCode string // 短信模板代码
    Enabled     bool   // 是否启用真实发送（false 时使用模拟）
}
```

在 `.env` 文件中配置：

```env
SMS_PROVIDER=aliyun
SMS_ACCESS_KEY_ID=your_access_key_id
SMS_ACCESS_KEY_SECRET=your_access_key_secret
SMS_SIGN_NAME=企业博客
SMS_TEMPLATE_CODE=SMS_123456789
SMS_ENABLED=true
```

### 3. 实现短信发送接口

创建 `internal/services/sms_provider.go`：

```go
package services

import (
    "fmt"
    "github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
    "github.com/aliyun/alibaba-cloud-sdk-go/sdk"
)

// SMSProvider 短信服务商接口
type SMSProvider interface {
    SendCode(phone, code string) error
}

// AliyunSMSProvider 阿里云短信服务商实现
type AliyunSMSProvider struct {
    client      *dysmsapi.Client
    signName    string
    templateCode string
}

func NewAliyunSMSProvider(accessKeyID, accessKeySecret, signName, templateCode string) (*AliyunSMSProvider, error) {
    client, err := dysmsapi.NewClientWithAccessKey("cn-hangzhou", accessKeyID, accessKeySecret)
    if err != nil {
        return nil, fmt.Errorf("failed to create aliyun sms client: %w", err)
    }
    
    return &AliyunSMSProvider{
        client:       client,
        signName:     signName,
        templateCode: templateCode,
    }, nil
}

func (p *AliyunSMSProvider) SendCode(phone, code string) error {
    request := dysmsapi.CreateSendSmsRequest()
    request.Scheme = "https"
    request.PhoneNumbers = phone
    request.SignName = p.signName
    request.TemplateCode = p.templateCode
    request.TemplateParam = fmt.Sprintf(`{"code":"%s"}`, code)
    
    response, err := p.client.SendSms(request)
    if err != nil {
        return fmt.Errorf("failed to send sms: %w", err)
    }
    
    if response.Code != "OK" {
        return fmt.Errorf("sms send failed: %s - %s", response.Code, response.Message)
    }
    
    return nil
}
```

### 4. 修改 SMSService

修改 `internal/services/sms_service.go` 中的 `SendCode` 方法：

```go
func (s *SMSService) SendCode(phone string) error {
    // ... 前面的防刷和验证码生成逻辑保持不变 ...
    
    // 根据配置决定使用真实发送还是模拟
    if config.AppConfig.SMS.Enabled && s.smsProvider != nil {
        // 使用真实短信服务商发送
        if err := s.smsProvider.SendCode(phone, code); err != nil {
            l := logger.GetLogger()
            l.Error().Err(err).Str("phone", phone).Msg("Failed to send SMS via provider")
            // 可以选择回退到模拟模式或返回错误
            return fmt.Errorf("failed to send SMS: %w", err)
        }
        l := logger.GetLogger()
        l.Info().Str("phone", phone).Msg("SMS code sent via provider")
    } else {
        // 模拟模式（开发/测试环境）
        l := logger.GetLogger()
        l.Info().
            Str("phone", phone).
            Str("code", code).
            Msg("SMS code sent (simulated)")
    }
    
    return nil
}
```

### 5. 初始化 SMS Provider

在 `cmd/server/main.go` 中初始化：

```go
// 初始化短信服务商（如果启用）
var smsProvider services.SMSProvider
if config.AppConfig.SMS.Enabled {
    switch config.AppConfig.SMS.Provider {
    case "aliyun":
        provider, err := services.NewAliyunSMSProvider(
            config.AppConfig.SMS.AccessKeyID,
            config.AppConfig.SMS.AccessKeySecret,
            config.AppConfig.SMS.SignName,
            config.AppConfig.SMS.TemplateCode,
        )
        if err != nil {
            l := logger.GetLogger()
            l.Warn().Err(err).Msg("Failed to init SMS provider, using simulated mode")
        } else {
            smsProvider = provider
        }
    // 可以添加其他服务商的初始化
    }
}

smsService := services.NewSMSService(smsRepo, userRepo)
if smsProvider != nil {
    smsService.SetProvider(smsProvider)
}
```

## 接入步骤总结

1. **选择服务商**：根据需求选择适合的短信服务商
2. **注册账号**：在服务商平台注册账号并获取 API 密钥
3. **申请签名和模板**：在服务商平台申请短信签名和模板
4. **安装 SDK**：安装对应服务商的 Go SDK
5. **实现接口**：实现 `SMSProvider` 接口
6. **配置环境变量**：添加服务商配置到 `.env` 文件
7. **测试验证**：在测试环境验证短信发送功能
8. **生产部署**：确保生产环境配置正确

## 注意事项

1. **成本控制**：
   - 设置合理的发送频率限制（已实现：1分钟1次）
   - 监控短信发送量，避免异常消耗
   - 考虑使用短信包或按量计费

2. **安全性**：
   - API 密钥不要提交到代码仓库
   - 使用环境变量或密钥管理服务
   - 定期轮换 API 密钥

3. **可靠性**：
   - 实现重试机制（服务商 SDK 通常已包含）
   - 监控发送成功率
   - 设置告警机制

4. **合规性**：
   - 遵守相关法律法规
   - 获取用户同意（隐私政策）
   - 不要发送营销短信（除非用户同意）

## 测试建议

1. **开发环境**：使用模拟模式（`SMS_ENABLED=false`），验证码在日志中查看
2. **测试环境**：可以接入服务商的测试账号或使用真实账号但限制发送量
3. **生产环境**：必须使用真实服务商，并监控发送情况

## 相关文件

- `internal/services/sms_service.go` - 短信服务实现
- `internal/repository/sms_repository.go` - 短信验证码数据访问层
- `internal/handlers/user_handler.go` - 短信验证码 API 处理
- `internal/models/user.go` - 用户模型（包含手机号字段）

## 参考资源

- [阿里云短信服务文档](https://help.aliyun.com/product/44282.html)
- [腾讯云短信服务文档](https://cloud.tencent.com/document/product/382)
- [Twilio SMS API 文档](https://www.twilio.com/docs/sms)

