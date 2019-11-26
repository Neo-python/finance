// 原作者:https://github.com/Wangjiaxing123/JwtDemo
package jwt_auth

import (
	"errors"
	plugins_pkg "finance/plugins"
	plugins "finance/plugins/common"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"time"
)

// 检查与校验token
func checkToken(context *gin.Context) {
	token := context.Request.Header.Get("Authorization")
	if token == "" {
		plugins.ApiExport(context).Error(4000, "未能从请求中获取用户身份信息.无法确认用户身份", errors.New("请求未携带token，无权限访问"))
		return
	}
	j := NewJWT()
	// parseToken 解析token包含的信息
	claims, err := j.ParseToken(token)
	if err != nil {
		if err == TokenExpired {

			plugins.ApiExport(context).Error(4001, "授权已过期", err)
			return

		}
		plugins.ApiExport(context).Error(4002, "token错误,无法验证用户身份信息.", err)
		return
	}
	if claims.AuthToken() != true {
		plugins.ApiExport(context).Error(4003, "授权已过期", errors.New("授权已过期"))
		return
	}
	// 继续交由下一个路由处理,并将解析出的信息传递下去
	context.Set("claims", claims)
}

// JWTAuth 中间件，检查token
func JWTAuth() gin.HandlerFunc {
	return checkToken
}

// JWT 签名结构
type JWT struct {
	SigningKey []byte
}

// 一些常量
var (
	TokenExpired     error = errors.New("Token is expired")
	TokenNotValidYet error = errors.New("Token not active yet")
	TokenMalformed   error = errors.New("That's not even a token")
	TokenInvalid     error = errors.New("Couldn't handle this token:")
)

// 载荷，可以加一些自己需要的信息
type CustomClaims struct {
	ID    string `json:"userId"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Iat   string `json:"iat"`
	jwt.StandardClaims
}

// 新建一个jwt实例
func NewJWT() *JWT {
	return &JWT{
		[]byte(GetSignKey()),
	}
}

// 获取signKey
func GetSignKey() string {
	return plugins_pkg.Config.JWTSecretKey
}

// CreateToken 生成一个token
func (j *JWT) CreateToken(claims *CustomClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.SigningKey)
}

// 解析Tokne
func (j *JWT) ParseToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.SigningKey, nil
	})
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, TokenMalformed
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				// Token is expired
				return nil, TokenExpired
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				return nil, TokenNotValidYet
			} else {
				return nil, TokenInvalid
			}
		}
	}
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, TokenInvalid
}

// 更新token
func (j *JWT) RefreshToken(tokenString string) (string, error) {
	jwt.TimeFunc = func() time.Time {
		return time.Unix(0, 0)
	}
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.SigningKey, nil
	})
	if err != nil {
		return "", err
	}
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		jwt.TimeFunc = time.Now
		claims.StandardClaims.ExpiresAt = time.Now().Add(1 * time.Hour).Unix()
		return j.CreateToken(claims)
	}
	return "", TokenInvalid
}

//
//func main() {
//	jwt_obj := NewJWT()
//	token, err := jwt_obj.CreateToken(&CustomClaims{})
//	fmt.Println(token, err)
//	fmt.Println(jwt_obj.ParseToken("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiIiLCJuYW1lIjoiIiwicGhvbmUiOiIifQ.YHZ7UIdo8V9SViG_1PBa8tf4-4DVRZttsfIT73iWM00"))
//}
