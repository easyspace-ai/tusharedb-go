package indicators

import (
	"math"
)

// Kline 单条K线数据
type Kline struct {
	Date   string
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
	Amount float64
}

// MAData 移动平均线数据
type MAData struct {
	MA5  *float64 `json:"ma5,omitempty"`
	MA10 *float64 `json:"ma10,omitempty"`
	MA20 *float64 `json:"ma20,omitempty"`
	MA60 *float64 `json:"ma60,omitempty"`
}

// MACDData MACD指标数据
type MACDData struct {
	DIF  float64 `json:"dif"`
	DEA  float64 `json:"dea"`
	MACD float64 `json:"macd"`
}

// BOLLData 布林带数据
type BOLLData struct {
	Upper float64 `json:"upper"`
	Mid   float64 `json:"mid"`
	Lower float64 `json:"lower"`
}

// KDJData KDJ指标数据
type KDJData struct {
	K float64 `json:"k"`
	D float64 `json:"d"`
	J float64 `json:"j"`
}

// RSIData RSI指标数据
type RSIData struct {
	RSI6  *float64 `json:"rsi6,omitempty"`
	RSI12 *float64 `json:"rsi12,omitempty"`
	RSI24 *float64 `json:"rsi24,omitempty"`
}

// Options 计算选项
type Options struct {
	MA   bool
	MACD bool
	BOLL bool
	KDJ  bool
	RSI  bool
}

// Result 计算结果
type Result struct {
	MA   *MAData   `json:"ma,omitempty"`
	MACD *MACDData `json:"macd,omitempty"`
	BOLL *BOLLData `json:"boll,omitempty"`
	KDJ  *KDJData  `json:"kdj,omitempty"`
	RSI  *RSIData  `json:"rsi,omitempty"`
}

// Calculate 计算技术指标
func Calculate(klines []Kline, opts Options) []Result {
	n := len(klines)
	if n == 0 {
		return nil
	}

	results := make([]Result, n)

	// 提取收盘价序列
	closes := make([]float64, n)
	for i, k := range klines {
		closes[i] = k.Close
	}

	// 计算MA
	if opts.MA {
		ma5 := SMA(closes, 5)
		ma10 := SMA(closes, 10)
		ma20 := SMA(closes, 20)
		ma60 := SMA(closes, 60)

		for i := 0; i < n; i++ {
			results[i].MA = &MAData{}
			if i >= 4 {
				results[i].MA.MA5 = &ma5[i]
			}
			if i >= 9 {
				results[i].MA.MA10 = &ma10[i]
			}
			if i >= 19 {
				results[i].MA.MA20 = &ma20[i]
			}
			if i >= 59 {
				results[i].MA.MA60 = &ma60[i]
			}
		}
	}

	// 计算MACD
	if opts.MACD {
		dif, dea, macd := MACD(closes, 12, 26, 9)
		for i := 0; i < n; i++ {
			results[i].MACD = &MACDData{
				DIF:  dif[i],
				DEA:  dea[i],
				MACD: macd[i],
			}
		}
	}

	// 计算BOLL
	if opts.BOLL {
		upper, mid, lower := BOLL(closes, 20, 2.0)
		for i := 0; i < n; i++ {
			results[i].BOLL = &BOLLData{
				Upper: upper[i],
				Mid:   mid[i],
				Lower: lower[i],
			}
		}
	}

	// 计算KDJ
	if opts.KDJ {
		k, d, j := KDJ(klines, 9, 3, 3)
		for i := 0; i < n; i++ {
			results[i].KDJ = &KDJData{
				K: k[i],
				D: d[i],
				J: j[i],
			}
		}
	}

	// 计算RSI
	if opts.RSI {
		rsi6 := RSI(closes, 6)
		rsi12 := RSI(closes, 12)
		rsi24 := RSI(closes, 24)
		for i := 0; i < n; i++ {
			results[i].RSI = &RSIData{}
			if !math.IsNaN(rsi6[i]) {
				results[i].RSI.RSI6 = &rsi6[i]
			}
			if !math.IsNaN(rsi12[i]) {
				results[i].RSI.RSI12 = &rsi12[i]
			}
			if !math.IsNaN(rsi24[i]) {
				results[i].RSI.RSI24 = &rsi24[i]
			}
		}
	}

	return results
}

// SMA 简单移动平均线
func SMA(data []float64, period int) []float64 {
	n := len(data)
	result := make([]float64, n)
	
	if n < period {
		for i := 0; i < n; i++ {
			result[i] = math.NaN()
		}
		return result
	}

	// 计算第一个值
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += data[i]
	}
	result[period-1] = sum / float64(period)

	// 使用滑动窗口计算剩余值
	for i := period; i < n; i++ {
		sum = sum - data[i-period] + data[i]
		result[i] = sum / float64(period)
	}

	// 前period-1个值为NaN
	for i := 0; i < period-1; i++ {
		result[i] = math.NaN()
	}

	return result
}

// EMA 指数移动平均线
func EMA(data []float64, period int) []float64 {
	n := len(data)
	result := make([]float64, n)
	
	if n == 0 {
		return result
	}

	multiplier := 2.0 / float64(period+1)
	
	// 第一个EMA使用SMA
	result[0] = data[0]
	sum := data[0]
	count := 1
	
	for i := 1; i < n; i++ {
		if count < period {
			sum += data[i]
			result[i] = sum / float64(count+1)
			count++
		} else {
			result[i] = (data[i]-result[i-1])*multiplier + result[i-1]
		}
	}
	
	return result
}

// MACD 计算MACD指标
func MACD(data []float64, fastPeriod, slowPeriod, signalPeriod int) (dif, dea, macd []float64) {
	n := len(data)
	dif = make([]float64, n)
	dea = make([]float64, n)
	macd = make([]float64, n)
	
	if n == 0 {
		return
	}

	// 计算快速和慢速EMA
	emaFast := EMA(data, fastPeriod)
	emaSlow := EMA(data, slowPeriod)
	
	// 计算DIF
	for i := 0; i < n; i++ {
		dif[i] = emaFast[i] - emaSlow[i]
	}
	
	// 计算DEA (DIF的EMA)
	dea = EMA(dif, signalPeriod)
	
	// 计算MACD柱
	for i := 0; i < n; i++ {
		macd[i] = (dif[i] - dea[i]) * 2
	}
	
	return
}

// BOLL 计算布林带
func BOLL(data []float64, period int, stdDev float64) (upper, mid, lower []float64) {
	n := len(data)
	upper = make([]float64, n)
	mid = make([]float64, n)
	lower = make([]float64, n)
	
	if n == 0 {
		return
	}

	// 中轨是SMA
	mid = SMA(data, period)
	
	for i := period - 1; i < n; i++ {
		// 计算标准差
		variance := 0.0
		for j := i - period + 1; j <= i; j++ {
			diff := data[j] - mid[i]
			variance += diff * diff
		}
		std := math.Sqrt(variance / float64(period))
		
		upper[i] = mid[i] + stdDev*std
		lower[i] = mid[i] - stdDev*std
	}
	
	// 前period-1个值填充为NaN
	for i := 0; i < period-1; i++ {
		upper[i] = math.NaN()
		lower[i] = math.NaN()
	}
	
	return
}

// KDJ 计算KDJ指标
func KDJ(klines []Kline, n, m1, m2 int) (k, d, j []float64) {
	length := len(klines)
	k = make([]float64, length)
	d = make([]float64, length)
	j = make([]float64, length)
	
	if length == 0 {
		return
	}

	// RSV = (收盘价 - N日内最低价) / (N日内最高价 - N日内最低价) * 100
	// K = (M1-1)/M1 * 昨日K + 1/M1 * 当日RSV
	// D = (M2-1)/M2 * 昨日D + 1/M2 * 当日K
	// J = 3*K - 2*D
	
	k[0] = 50
	d[0] = 50
	j[0] = 50
	
	for i := 1; i < length; i++ {
		start := 0
		if i >= n {
			start = i - n + 1
		}
		
		// 计算N日内的最高和最低价
		high := klines[start].High
		low := klines[start].Low
		for t := start; t <= i; t++ {
			if klines[t].High > high {
				high = klines[t].High
			}
			if klines[t].Low < low {
				low = klines[t].Low
			}
		}
		
		// 计算RSV
		var rsv float64
		if high != low {
			rsv = (klines[i].Close - low) / (high - low) * 100
		} else {
			rsv = 50
		}
		
		// 计算K, D, J
		k[i] = (float64(m1-1)/float64(m1))*k[i-1] + (1.0/float64(m1))*rsv
		d[i] = (float64(m2-1)/float64(m2))*d[i-1] + (1.0/float64(m2))*k[i]
		j[i] = 3*k[i] - 2*d[i]
	}
	
	return
}

// RSI 计算RSI指标
func RSI(data []float64, period int) []float64 {
	n := len(data)
	result := make([]float64, n)
	
	if n == 0 || period <= 0 {
		return result
	}

	// 初始化为NaN
	for i := 0; i < n; i++ {
		result[i] = math.NaN()
	}

	// 计算价格变动
	changes := make([]float64, n)
	for i := 1; i < n; i++ {
		changes[i] = data[i] - data[i-1]
	}

	// 计算初始平均涨跌
	avgGain := 0.0
	avgLoss := 0.0
	count := 0
	for i := 1; i <= period && i < n; i++ {
		if changes[i] > 0 {
			avgGain += changes[i]
		} else {
			avgLoss += -changes[i]
		}
		count++
	}
	
	if count > 0 {
		avgGain /= float64(period)
		avgLoss /= float64(period)
	}

	// 计算第一个RSI
	if period < n {
		if avgLoss == 0 {
			result[period] = 100
		} else {
			rs := avgGain / avgLoss
			result[period] = 100 - (100 / (1 + rs))
		}
	}

	// 使用平滑方法计算后续RSI
	for i := period + 1; i < n; i++ {
		gain := 0.0
		loss := 0.0
		if changes[i] > 0 {
			gain = changes[i]
		} else {
			loss = -changes[i]
		}
		
		avgGain = (avgGain*float64(period-1) + gain) / float64(period)
		avgLoss = (avgLoss*float64(period-1) + loss) / float64(period)
		
		if avgLoss == 0 {
			result[i] = 100
		} else {
			rs := avgGain / avgLoss
			result[i] = 100 - (100 / (1 + rs))
		}
	}
	
	return result
}
