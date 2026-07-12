package service

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const tokenIssuer = "go-todo"

var ErrInvalidToken = errors.New("invalid token")

// 定义访问令牌（Access Token）的声明结构体，包含自定义的会话ID（sid）以及JWT的预定义声明。
type AccessTokenClaims struct {
	SessionID            string `json:"sid"`
	jwt.RegisteredClaims        //匿名嵌入另一个结构体，jwt的预定义声明结构体
}

// 生成访问令牌（Access Token），包含用户ID、会话ID、签发者密钥以及过期时间
func GenerateAccessToken(userID uint64, sessionID string, secret string, ttl time.Duration) (string, error) {
	//记录当前时间
	now := time.Now()
	// 创建访问令牌的声明（Claims）
	claims := AccessTokenClaims{
		SessionID: sessionID, //自定义的sid
		RegisteredClaims: jwt.RegisteredClaims{ //jwt的预定义声明
			Subject:   strconv.FormatUint(userID, 10),   //将用户ID转换为字符串作为主题
			Issuer:    tokenIssuer,                      //签发者
			IssuedAt:  jwt.NewNumericDate(now),          //签发时间
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)), //过期时间
		},
	}
	// 创建一个新的JWT令牌，使用HS256签名方法，并将声明附加到令牌中
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// 使用提供的密钥对令牌进行签名，并返回生成的访问令牌字符串
	return token.SignedString([]byte(secret))
}

// 解析访问令牌（Access Token），验证其有效性并提取声明信息
func ParseAccessToken(tokenString string, secret string) (*AccessTokenClaims, error) {
	claims := &AccessTokenClaims{}
	// 使用 jwt.ParseWithClaims 解析访问令牌，并验证签名和声明
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (any, error) {
			return []byte(secret), nil // 返回用于验证签名的密钥
		},
		jwt.WithValidMethods([]string{ // 指定有效的签名方法
			jwt.SigningMethodHS256.Alg(),
		}),
		jwt.WithIssuer(tokenIssuer),  // 指定签发者
		jwt.WithExpirationRequired(), // 指定必须包含过期时间
		jwt.WithIssuedAt(),           // 指定必须包含签发时间
	)
	// 检查解析过程中是否出现错误，或者令牌是否无效
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	// 检查声明中的主题、会话ID和签发时间是否为空
	if claims.Subject == "" ||
		claims.SessionID == "" ||
		claims.IssuedAt == nil {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// 生成刷新令牌（Refresh Token），用于获取新的访问令牌
func GenerateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	// 使用加密随机数生成器生成32字节的随机数据
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	// 将生成的随机字节数据进行Base64编码，并返回作为刷新令牌字符串
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
