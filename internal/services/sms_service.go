// Package services 提供业务逻辑层的服务实现
package services

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"enterprise-blog/internal/database"
	"enterprise-blog/internal/models"
	"enterprise-blog/internal/repository"
	"enterprise-blog/pkg/logger"
)

// SMSService 短信服务，提供短信验证码相关的业务逻辑
type SMSService struct {
	smsRepo  *repository.SMSRepository
	userRepo *repository.UserRepository
	jwtMgr   interface{} // 占位，实际需要 JWTManager
}

// NewSMSService 创建新的短信服务实例
// smsRepo: 短信验证码数据访问层仓库
// userRepo: 用户数据访问层仓库，用于查找或创建用户
func NewSMSService(smsRepo *repository.SMSRepository, userRepo *repository.UserRepository) *SMSService {
	return &SMSService{
		smsRepo:  smsRepo,
		userRepo: userRepo,
	}
}

// SetJWTManager 设置 JWT 管理器（在初始化时调用）
func (s *SMSService) SetJWTManager(jwtMgr interface{}) {
	s.jwtMgr = jwtMgr
}

// LoginWithPhone 使用手机号和验证码登录，返回 token 和用户信息
func (s *SMSService) LoginWithPhone(phone, code string) (string, *models.User, error) {
	user, err := s.VerifyCode(phone, code)
	if err != nil {
		return "", nil, err
	}

	// 检查用户状态
	if user.Status != "active" {
		return "", nil, errors.New("user account is not active")
	}

	// 生成 JWT token（需要 JWTManager，这里先返回错误，由 handler 层处理）
	// 实际应该在 handler 层调用 jwtMgr.GenerateToken
	return "", user, nil
}

/**
 * @function SendCode
 * @description 发送短信验证码。
 *
 * @param {string} phone - 手机号码
 *
 * @returns {error} 如果发送失败则返回错误
 *
 * @remarks
 * - **当前实现**：模拟实现，仅在日志中输出验证码（开发/测试环境）
 * - **生产环境**：必须接入真实的短信服务商（如阿里云、腾讯云等）
 * - **防刷机制**：1分钟内只能发送一次验证码
 * - **验证码有效期**：5分钟
 * - **存储方式**：同时存储到数据库和 Redis（Redis 用于快速验证）
 *
 * @todo
 * - 接入短信服务商 API（阿里云、腾讯云、Twilio 等）
 * - 实现服务商接口抽象，支持多种服务商切换
 * - 添加发送失败重试机制
 * - 添加发送量监控和告警
 *
 * @see
 * - [短信接入指南](../../docs/SMS_INTEGRATION.md) - 详细的接入步骤和示例代码
 *
 * @interview_points
 * - 为什么需要接入短信服务商？（合规、用户体验、安全性）
 * - 如何实现服务商切换？（接口抽象，依赖注入）
 * - 如何防止验证码被刷？（频率限制、IP限制）
 * - 验证码存储在哪里？（数据库持久化，Redis 加速验证）
 */
func (s *SMSService) SendCode(phone string) error {
	// 检查最近1分钟内是否已发送过验证码（防刷）
	oneMinAgo := time.Now().Add(-1 * time.Minute)
	count, err := s.smsRepo.GetRecentCodeCount(phone, oneMinAgo)
	if err != nil {
		return fmt.Errorf("failed to check recent codes: %w", err)
	}
	if count >= 1 {
		return errors.New("验证码发送过于频繁，请稍后再试")
	}

	// 生成6位数字验证码
	code := fmt.Sprintf("%06d", rand.Intn(1000000))

	// 验证码5分钟有效
	expiresAt := time.Now().Add(5 * time.Minute)

	smsCode := &models.SMSCode{
		Phone:     phone,
		Code:      code,
		Used:      false,
		ExpiresAt: expiresAt,
	}

	if err := s.smsRepo.Create(smsCode); err != nil {
		return fmt.Errorf("failed to save sms code: %w", err)
	}

	// 同时存储到 Redis（如果可用），用于快速验证
	if database.RedisClient != nil {
		key := fmt.Sprintf("sms:code:%s", phone)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = database.RedisClient.Set(ctx, key, code, 5*time.Minute).Err()
	}

	// TODO: 实际生产环境应调用短信服务商 API 发送短信
	// 当前为模拟实现，仅在日志中输出验证码（开发/测试环境）
	l := logger.GetLogger()
	l.Info().
		Str("phone", phone).
		Str("code", code).
		Msg("SMS code sent (simulated)")

	return nil
}

// VerifyCode 验证验证码并返回用户（如果存在则返回，不存在则自动创建）
func (s *SMSService) VerifyCode(phone, code string) (*models.User, error) {
	// 先尝试从 Redis 快速验证
	if database.RedisClient != nil {
		key := fmt.Sprintf("sms:code:%s", phone)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		storedCode, err := database.RedisClient.Get(ctx, key).Result()
		if err == nil && storedCode == code {
			// Redis 验证通过，标记数据库中的验证码为已使用
			smsCode, err := s.smsRepo.GetValidCode(phone, code)
			if err == nil {
				_ = s.smsRepo.MarkAsUsed(smsCode.ID)
			}
			_ = database.RedisClient.Del(ctx, key).Err()

			// 查找或创建用户
			return s.findOrCreateUser(phone)
		}
	}

	// 从数据库验证
	smsCode, err := s.smsRepo.GetValidCode(phone, code)
	if err != nil {
		return nil, errors.New("验证码无效或已过期")
	}

	// 标记为已使用
	if err := s.smsRepo.MarkAsUsed(smsCode.ID); err != nil {
		return nil, fmt.Errorf("failed to mark code as used: %w", err)
	}

	// 删除 Redis 中的验证码
	if database.RedisClient != nil {
		key := fmt.Sprintf("sms:code:%s", phone)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = database.RedisClient.Del(ctx, key).Err()
	}

	// 查找或创建用户
	return s.findOrCreateUser(phone)
}

// findOrCreateUser 根据手机号查找用户，不存在则自动创建
func (s *SMSService) findOrCreateUser(phone string) (*models.User, error) {
	user, err := s.userRepo.GetByPhone(phone)
	if err == nil {
		// 用户已存在
		return user, nil
	}

	// 用户不存在，自动创建
	user = &models.User{
		Phone:    phone,
		Username: fmt.Sprintf("user_%s", phone[len(phone)-4:]), // 使用手机号后4位作为默认用户名
		Email:    fmt.Sprintf("%s@phone.local", phone),          // 生成临时邮箱
		Password: "", // 手机号登录不需要密码
		Role:     models.RoleReader,
		Status:   "active",
	}

	// 生成一个随机密码（虽然不会用到，但数据库字段要求非空）
	user.Password = fmt.Sprintf("phone_%s_%d", phone, time.Now().Unix())
	if err := user.HashPassword(); err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// 重新获取完整用户信息
	return s.userRepo.GetByID(user.ID)
}

