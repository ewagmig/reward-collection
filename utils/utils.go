package utils

import (
	"crypto/sha512"
	"encoding/base64"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/spf13/cast"
)

var rnd = rand.NewSource(time.Now().UnixNano())

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// StrInSlice checks if t is in ss slice.
func StrInSlice(ss []string, t string) bool {
	for _, s := range ss {
		if s == t {
			return true
		}
	}
	return false
}

// GetCommonFilterQueryParams returns the common filter query and arguments
func GetCommonFilterQueryParams(ctx *gin.Context) (query string, args []interface{}) {
	// start_date
	if s, e := ctx.GetQuery("start_date"); e {
		date, err := cast.StringToDate(s)
		if err == nil {
			query += " start_date >= ? and"
			args = append(args, date)
		}
	}

	// end_date
	if s, e := ctx.GetQuery("end_date"); e {
		date, err := cast.StringToDate(s)
		if err == nil {
			query += " end_date < ? and"
			args = append(args, date)
		}
	}

	if query != "" {
		query = query[:len(query)-3] // truncate the last 'and'
	}

	return
}

// GetPagingQueryParm returns params about paging.
func GetPagingQueryParm(ctx *gin.Context, orderBy ...string) (pageSize, pageIndex uint, unlimited bool, orderby string) {
	var err error
	ps := ctx.Query("page-size")
	if pageSize, err = cast.ToUintE(ps); err != nil {
		pageSize = 10 // default
	}

	pi := ctx.Query("page-index")
	if pageIndex, err = cast.ToUintE(pi); err != nil {
		pageIndex = 0 // default
	} else {
		pageIndex = pageIndex - 1
	}

	u := ctx.Query("unlimited")
	if unlimited, err = cast.ToBoolE(u); err != nil {
		unlimited = false // default
	}

	var defaultOrder = "updated_at DESC"
	if len(orderBy) > 0 {
		defaultOrder = orderBy[0]
	}
	orderby = ctx.DefaultQuery("orderby", defaultOrder)

	return
}

// GetBoolParam returns the boolean value with the specified key.
func GetBoolParam(ctx *gin.Context, key string) (b bool, err error) {
	v := ctx.Param(key)
	return cast.ToBoolE(v)
}

// GetIntParam returns the int value with the specified key.
func GetIntParam(ctx *gin.Context, key string) (i int, err error) {
	v := ctx.Param(key)
	return cast.ToIntE(v)
}

// GetStringParam returns the string value withe the specified key.
func GetStringParam(ctx *gin.Context, key string) string {
	return ctx.Param(key)
}

// GetJSONBody returns the JSON object from the request body.
func GetJSONBody(ctx *gin.Context, out interface{}) error {
	return binding.JSON.Bind(ctx.Request, out)
}

// WriteSuccess return a successfull response.
func WriteSuccess(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "success",
	})
}

func WritePagination(ctx *gin.Context, data interface{}, size, index, total uint, unlimited bool) {
	ctx.JSON(http.StatusOK, gin.H{
		"data": data,
		"paging": gin.H{
			"total_count": total,
			"page_count":  (total + size - 1) / size,
			"page_index":  index + 1,
			"page_size":   size,
			"unlimited":   unlimited,
		},
	})
}

// NewValueTypeOf creates a new object with same type of v.
func NewValueTypeOf(v interface{}) interface{} {
	vt := reflect.TypeOf(v)
	switch vt.Kind() {
	case reflect.Ptr:
		vt = vt.Elem()
	default:
	}

	return reflect.New(vt).Interface()
}

// NewValueSliceTypeOf creates a new slice pointer with element type of v.
// For example, if v is int, *[]int will be returned.
func NewValueSliceTypeOf(v interface{}) interface{} {
	mt := reflect.TypeOf(v)
	switch mt.Kind() {
	case reflect.Ptr:
		mt = mt.Elem()
	default:
	}

	slice := reflect.MakeSlice(reflect.SliceOf(mt), 0, 1)
	slicep := reflect.New(slice.Type())
	slicep.Elem().Set(slice)

	return slicep.Interface()
}

func CalcSha384Hash(in []byte) ([]byte, error) {
	var sha = sha512.New384()
	_, err := sha.Write(in)
	if err != nil {
		return []byte{}, err
	}

	return sha.Sum([]byte(nil)), nil
}

// IsBase64 returns true if s is valid base64 format string.
func IsBase64(s string) bool {
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}

func StrToUintSlice(s string) []uint {
	var r []uint
	for _, s := range strings.Split(s, ",") {
		u, err := cast.ToUintE(s)
		if err != nil {
			continue
		}

		r = append(r, u)
	}

	return r
}

func StrToStrSlice(s string) []string {
	var r []string
	for _, s := range strings.Split(s, ",") {
		r = append(r, s)
	}
	return r
}

func StrSliceToBytesSlice(src []string) [][]byte {
	var dest [][]byte
	for _, s := range src {
		dest = append(dest, []byte(s))
	}

	return dest
}

func Base64StrSliceToBytesSlice(src []string) ([][]byte, error) {
	var dest [][]byte
	for _, s := range src {
		sb, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			return dest, err
		}

		dest = append(dest, sb)
	}

	return dest, nil
}

func ShuffleStrSlice(a []string) []string {
	n := len(a)
	returnedSlice := make([]string, n)
	rand.Seed(time.Now().UnixNano())
	indices := rand.Perm(n)
	for i, idx := range indices {
		returnedSlice[i] = a[idx]
	}
	return returnedSlice
}

func UintInArray(t uint, arr []uint) bool {
	for _, a := range arr {
		if t == a {
			return true
		}
	}

	return false
}

func UintArrayDiff(arr1, arr2 []uint) []uint {
	ret := []uint{}
	for _, a := range arr1 {
		var ok bool
		for _, b := range arr2 {
			if a == b {
				ok = true
				break
			}
		}
		if !ok {
			ret = append(ret, a)
		}
	}
	return ret
}

func StringArrayDiff(arr1, arr2 []string) []string {
	ret := []string{}
	for _, a := range arr1 {
		var ok bool
		for _, b := range arr2 {
			if a == b {
				ok = true
				break
			}
		}
		if !ok {
			ret = append(ret, a)
		}
	}
	return ret
}

// B64Encode base64 encodes bytes
func B64Encode(buf []byte) string {
	return base64.StdEncoding.EncodeToString(buf)
}

// B64Decode base64 decodes a string
func B64Decode(str string) (buf []byte, err error) {
	return base64.StdEncoding.DecodeString(str)
}

// RandomString returns a random string
func RandomString(n int) string {
	b := make([]byte, n)

	for i, cache, remain := n-1, rnd.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rnd.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

// RemoveDuplicateStrings remove duplicate string from elements.
func RemoveDuplicateStrings(elements []string) []string {
	founds := map[string]bool{}

	// Create a map of all unique elements.
	for v := range elements {
		founds[elements[v]] = true
	}

	// Place all keys from the map into a slice.
	result := []string{}
	for key := range founds {
		result = append(result, key)
	}
	return result
}

// RemoveDuplicateUint removes duplicate uint from elements
func RemoveDuplicateUints(elements []uint) []uint {
	founds := map[uint]bool{}

	// Create a map of all unique elements.
	for v := range elements {
		founds[elements[v]] = true
	}

	// Place all keys from the map into a slice.
	result := []uint{}
	for key := range founds {
		result = append(result, key)
	}
	return result
}

func GetDir(path string) string {
	return getDir(path)
}

func getDir(path string) string {
	return subString(path, 0, strings.LastIndex(path, "/"))
}

func subString(str string, start, end int) string {
	rs := []rune(str)
	length := len(rs)

	if start < 0 || start > length {
		panic("start is wrong")
	}

	if end < start || end > length {
		panic("end is wrong")
	}

	return string(rs[start:end])
}

// 判断所给路径文件/文件夹是否存在
func PathExists(path string) bool {
	return pathExists(path)
}

// 判断所给路径文件/文件夹是否存在
func pathExists(path string) bool {
	//os.Stat获取文件信息
	_, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

// 判断所给路径是否为文件夹
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// 判断所给路径是否为文件
func IsFile(path string) bool {
	return !IsDir(path)
}

func WriteBlockBytes(ctx *gin.Context, blockBytes []byte) {
	ctx.Status(http.StatusOK)
	ctx.Writer.Header().Set("Content-Type", "application/octet-stream")
	ctx.Writer.Write(blockBytes)
}
