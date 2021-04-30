package manipulate

import (
	"math/rand"
	"reflect"
	"regexp"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

//StringWithCharset Makes random string based on lenght
func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// MaxWordsInSentences - return max words in one sentences
func MaxWordsInSentences(S string) (result int) {
	r, _ := regexp.Compile("[.||?||!]")
	count := strings.Count(S, ".") + strings.Count(S, "!") + strings.Count(S, "?") // Total sentaces

	for i := 0; i < count; i++ {
		sentence := r.Split(S, count)[i]
		splitSentence := strings.Split(sentence, " ")

		var R []string
		for _, str := range splitSentence {
			if str != "" {
				R = append(R, str)
			}
		}

		if len(R) > result {
			result = len(R)
		}
	}

	return
}

// DateInRange return if date is in 1 month ago or not
// Return true if date in within 1 month
// Return false is date is older that 1 month
func DateInRange(dateFormat, dateString string) (bool, error) {

	dateStamp, err := time.Parse(dateFormat, dateString)

	if err != nil {
		return false, err
	}

	log.Debugln("Date to check:", dateStamp)
	today := time.Now()

	oneMonthAgo := today.AddDate(0, -1, 0)
	log.Debugf("1 Month Ago was %s", oneMonthAgo.Format(time.RFC3339))

	if dateStamp.Before(oneMonthAgo) {
		return false, nil
	}

	return true, nil
}

// DateInRangeDays return if date is in x Days ago or not,
// Return true if date in within x Days,
// Return false is date is older that x Days,
func DateInRangeDays(dateCheck string, days int) (bool, error) {

	dateStamp, err := time.Parse(time.RFC3339, dateCheck)

	if err != nil {
		return false, err
	}

	log.Debugln("Date to check:", dateStamp)
	today := time.Now()

	daysAgo := today.AddDate(0, 0, -days)
	log.Debugf("Days Ago was %s", daysAgo.Format(time.RFC3339))

	if dateStamp.Before(daysAgo) {
		log.Infof("Days Ago was %s; return false", daysAgo.Format(time.RFC3339))
		return false, nil
	}

	return true, nil
}

// ItemExists check if 1 item exists in an array
func ItemExists(arrayType interface{}, item interface{}) bool {
	arr := reflect.ValueOf(arrayType)

	if arr.Kind() != reflect.Array {
		//panic("Invalid data-type")
		return false
	}

	for i := 0; i < arr.Len(); i++ {
		if arr.Index(i).Interface() == item {
			return true
		}
	}

	return false
}
