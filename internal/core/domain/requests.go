package domain

import (
	"fmt"
	"net/http"
	"time"
)

type RequestInfo struct {
	Lang          string  `json:"lang"`
	RequestID     string  `json:"request_id"`
	IpAddress     *string `json:"ip_address"`
	DeviceID      *string `json:"device_id"`
	Latitude      *string `json:"latitude"`
	Longitude     *string `json:"longitude"`
	UserAgent     *string `json:"user_agent"`
	Url           *string `json:"url"`
	Referrer      *string `json:"referrer"`
	SessionID     *string `json:"session_id"`
	AcceptLang    *string `json:"accept_language"`
	Platform      *string `json:"platform"`
	AppVersion    *string `json:"app_version"`
	Timezone      *string `json:"timezone"`
	RequestMethod *string `json:"request_method"`
	Host          *string `json:"host"`
	ForwardedFor  *string `json:"forwarded_for"`
	XRealIP       *string `json:"x_real_ip"`
	ContentType   *string `json:"content_type"`
	ContentLength *int64  `json:"content_length"`
	Cookies       *string `json:"cookies"`
	Connection    *string `json:"connection"`
	Origin        *string `json:"origin"`
}

func (rm RequestInfo) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"lang":            rm.Lang,
		"ip_address":      rm.IpAddress,
		"device_id":       rm.DeviceID,
		"latitude":        rm.Latitude,
		"longitude":       rm.Longitude,
		"user_agent":      rm.UserAgent,
		"url":             rm.Url,
		"referrer":        rm.Referrer,
		"session_id":      rm.SessionID,
		"accept_language": rm.AcceptLang,
		"platform":        rm.Platform,
		"app_version":     rm.AppVersion,
		"timezone":        rm.Timezone,
		"request_method":  rm.RequestMethod,
		"host":            rm.Host,
		"forwarded_for":   rm.ForwardedFor,
		"x_real_ip":       rm.XRealIP,
		"content_type":    rm.ContentType,
		"content_length":  rm.ContentLength,
		"cookies":         rm.Cookies,
		"request_id":      rm.RequestID,
		"connection":      rm.Connection,
		"origin":          rm.Origin,
	}
}

func (rm RequestInfo) GetLocation() string {
	if rm.Latitude != nil && rm.Longitude != nil {
		return fmt.Sprintf("%s,%s", *rm.Latitude, *rm.Longitude)
	}
	return ""
}

func (rm RequestInfo) GetDeviceId() string {
	if rm.DeviceID != nil {
		return *rm.DeviceID
	}
	return ""
}

func (rm RequestInfo) GetIpAddress() string {
	if rm.IpAddress != nil {
		return *rm.IpAddress
	}
	return ""
}

func (rm RequestInfo) ActivityLogMeta() LogActivityMetadata {
	return LogActivityMetadata{
		IP:       rm.GetIpAddress(),
		Device:   rm.GetDeviceId(),
		Location: rm.GetLocation(),
	}
}

// ExtractRequestInfo extracts RequestInfo from *http.Request
func ExtractRequestInfo(r *http.Request) RequestInfo {
	getHeader := func(key string) *string {
		if val := r.Header.Get(key); val != "" {
			return &val
		}
		return nil
	}

	var contentLength *int64
	if r.ContentLength > 0 {
		cl := r.ContentLength
		contentLength = &cl
	}

	var cookies *string
	if len(r.Cookies()) > 0 {
		cookieStr := ""
		for i, c := range r.Cookies() {
			if i > 0 {
				cookieStr += "; "
			}
			cookieStr += c.Name + "=" + c.Value
		}
		cookies = &cookieStr
	}

	urlStr := r.URL.String()
	method := r.Method
	host := r.Host

	lang := r.URL.Query().Get("lang")
	if lang == "" {
		langHeader := r.Header.Get("Accept-Language")
		if langHeader != "" {
			lang = langHeader
		} else {
			lang = "en"
		}
	}

	requestId := getHeader("X-Request-Id")
	if requestId == nil {
		generatedId := fmt.Sprintf("req-%d", time.Now().UnixNano())
		requestId = &generatedId
	}

	ip := getHeader("X-Forwarded-For")
	if ip == nil || *ip == "" {
		ip = getHeader("X-Real-Ip")
	}
	if ip == nil || *ip == "" {
		ip = getHeader("RemoteAddr")
	}
	if ip == nil || *ip == "" {
		ip = &r.RemoteAddr
	}

	return RequestInfo{
		Lang:          lang,
		IpAddress:     ip,
		DeviceID:      getHeader("X-Device-Id"),
		Latitude:      getHeader("X-Latitude"),
		Longitude:     getHeader("X-Longitude"),
		UserAgent:     getHeader("User-Agent"),
		Url:           &urlStr,
		Referrer:      getHeader("Referer"),
		SessionID:     getHeader("X-Session-Id"),
		AcceptLang:    getHeader("Accept-Language"),
		Platform:      getHeader("X-Platform"),
		AppVersion:    getHeader("X-App-Version"),
		Timezone:      getHeader("X-Timezone"),
		RequestMethod: &method,
		Host:          &host,
		ForwardedFor:  getHeader("X-Forwarded-For"),
		XRealIP:       getHeader("X-Real-Ip"),
		ContentType:   getHeader("Content-Type"),
		ContentLength: contentLength,
		Cookies:       cookies,
		RequestID:     *requestId,
		Connection:    getHeader("Connection"),
		Origin:        getHeader("Origin"),
	}
}
