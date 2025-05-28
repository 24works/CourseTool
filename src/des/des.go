package des

import (
	"fmt"
	"strconv"
)

// strEnc 函數實現了 DES 加密
// data: 要加密的字串
// firstKey, secondKey, thirdKey: 三個加密金鑰 (用於 3DES)
func StrEnc(data, firstKey, secondKey, thirdKey string) string {
	leng := len(data)
	encData := ""

	var firstKeyBt, secondKeyBt, thirdKeyBt [][]int
	var firstLength, secondLength, thirdLength int

	if firstKey != "" {
		firstKeyBt = getKeyBytes(firstKey)
		firstLength = len(firstKeyBt)
	}
	if secondKey != "" {
		secondKeyBt = getKeyBytes(secondKey)
		secondLength = len(secondKeyBt)
	}
	if thirdKey != "" {
		thirdKeyBt = getKeyBytes(thirdKey)
		thirdLength = len(thirdKeyBt)
	}

	if leng > 0 {
		if leng < 4 {
			bt := strToBt(data)
			var encByte []int

			if firstKey != "" && secondKey != "" && thirdKey != "" {
				tempBt := bt
				for x := 0; x < firstLength; x++ {
					tempBt = enc(tempBt, firstKeyBt[x])
				}
				for y := 0; y < secondLength; y++ {
					tempBt = enc(tempBt, secondKeyBt[y])
				}
				for z := 0; z < thirdLength; z++ {
					tempBt = enc(tempBt, thirdKeyBt[z])
				}
				encByte = tempBt
			} else if firstKey != "" && secondKey != "" {
				tempBt := bt
				for x := 0; x < firstLength; x++ {
					tempBt = enc(tempBt, firstKeyBt[x])
				}
				for y := 0; y < secondLength; y++ {
					tempBt = enc(tempBt, secondKeyBt[y])
				}
				encByte = tempBt
			} else if firstKey != "" {
				tempBt := bt
				for x := 0; x < firstLength; x++ {
					tempBt = enc(tempBt, firstKeyBt[x])
				}
				encByte = tempBt
			}
			encData = bt64ToHex(encByte)
		} else {
			iterator := leng / 4
			remainder := leng % 4

			for i := 0; i < iterator; i++ {
				tempData := data[i*4 : i*4+4]
				tempByte := strToBt(tempData)
				var encByte []int

				if firstKey != "" && secondKey != "" && thirdKey != "" {
					tempBt := tempByte
					for x := 0; x < firstLength; x++ {
						tempBt = enc(tempBt, firstKeyBt[x])
					}
					for y := 0; y < secondLength; y++ {
						tempBt = enc(tempBt, secondKeyBt[y])
					}
					for z := 0; z < thirdLength; z++ {
						tempBt = enc(tempBt, thirdKeyBt[z])
					}
					encByte = tempBt
				} else if firstKey != "" && secondKey != "" {
					tempBt := tempByte
					for x := 0; x < firstLength; x++ {
						tempBt = enc(tempBt, firstKeyBt[x])
					}
					for y := 0; y < secondLength; y++ {
						tempBt = enc(tempBt, secondKeyBt[y])
					}
					encByte = tempBt
				} else if firstKey != "" {
					tempBt := tempByte
					for x := 0; x < firstLength; x++ {
						tempBt = enc(tempBt, firstKeyBt[x])
					}
					encByte = tempBt
				}
				encData += bt64ToHex(encByte)
			}
			if remainder > 0 {
				remainderData := data[iterator*4 : leng]
				tempByte := strToBt(remainderData)
				var encByte []int

				if firstKey != "" && secondKey != "" && thirdKey != "" {
					tempBt := tempByte
					for x := 0; x < firstLength; x++ {
						tempBt = enc(tempBt, firstKeyBt[x])
					}
					for y := 0; y < secondLength; y++ {
						tempBt = enc(tempBt, secondKeyBt[y])
					}
					for z := 0; z < thirdLength; z++ {
						tempBt = enc(tempBt, thirdKeyBt[z])
					}
					encByte = tempBt
				} else if firstKey != "" && secondKey != "" {
					tempBt := tempByte
					for x := 0; x < firstLength; x++ {
						tempBt = enc(tempBt, firstKeyBt[x])
					}
					for y := 0; y < secondLength; y++ {
						tempBt = enc(tempBt, secondKeyBt[y])
					}
					encByte = tempBt
				} else if firstKey != "" {
					tempBt := tempByte
					for x := 0; x < firstLength; x++ {
						tempBt = enc(tempBt, firstKeyBt[x])
					}
					encByte = tempBt
				}
				encData += bt64ToHex(encByte)
			}
		}
	}
	return encData
}

// strDec 函數實現了 DES 解密
// data: 要解密的字串
// firstKey, secondKey, thirdKey: 三個解密金鑰 (用於 3DES)
func StrDec(data, firstKey, secondKey, thirdKey string) string {
	leng := len(data)
	decStr := ""

	var firstKeyBt, secondKeyBt, thirdKeyBt [][]int
	var firstLength, secondLength, thirdLength int

	if firstKey != "" {
		firstKeyBt = getKeyBytes(firstKey)
		firstLength = len(firstKeyBt)
	}
	if secondKey != "" {
		secondKeyBt = getKeyBytes(secondKey)
		secondLength = len(secondKeyBt)
	}
	if thirdKey != "" {
		thirdKeyBt = getKeyBytes(thirdKey)
		thirdLength = len(thirdKeyBt)
	}

	iterator := leng / 16
	for i := 0; i < iterator; i++ {
		tempData := data[i*16 : i*16+16]
		strByte := hexToBt64(tempData)
		intByte := make([]int, 64)
		for j := 0; j < 64; j++ {
			intByte[j], _ = strconv.Atoi(string(strByte[j]))
		}

		var decByte []int
		if firstKey != "" && secondKey != "" && thirdKey != "" {
			tempBt := intByte
			for x := thirdLength - 1; x >= 0; x-- {
				tempBt = dec(tempBt, thirdKeyBt[x])
			}
			for y := secondLength - 1; y >= 0; y-- {
				tempBt = dec(tempBt, secondKeyBt[y])
			}
			for z := firstLength - 1; z >= 0; z-- {
				tempBt = dec(tempBt, firstKeyBt[z])
			}
			decByte = tempBt
		} else if firstKey != "" && secondKey != "" {
			tempBt := intByte
			for x := secondLength - 1; x >= 0; x-- {
				tempBt = dec(tempBt, secondKeyBt[x])
			}
			for y := firstLength - 1; y >= 0; y-- {
				tempBt = dec(tempBt, firstKeyBt[y])
			}
			decByte = tempBt
		} else if firstKey != "" {
			tempBt := intByte
			for x := firstLength - 1; x >= 0; x-- {
				tempBt = dec(tempBt, firstKeyBt[x])
			}
			decByte = tempBt
		}
		decStr += byteToString(decByte)
	}
	return decStr
}

// getKeyBytes 將金鑰字串轉換為位元陣列
func getKeyBytes(key string) [][]int {
	keyBytes := make([][]int, 0)
	leng := len(key)
	iterator := leng / 4
	remainder := leng % 4

	for i := 0; i < iterator; i++ {
		keyBytes = append(keyBytes, strToBt(key[i*4:i*4+4]))
	}
	if remainder > 0 {
		keyBytes = append(keyBytes, strToBt(key[iterator*4:leng]))
	}
	return keyBytes
}

// strToBt 將長度 <= 4 的字串轉換為 64 位元陣列
func strToBt(str string) []int {
	leng := len(str)
	bt := make([]int, 64)
	if leng < 4 {
		for i := 0; i < leng; i++ {
			k := int(str[i])
			for j := 0; j < 16; j++ {
				pow := 1
				for m := 15; m > j; m-- {
					pow *= 2
				}
				bt[16*i+j] = (k / pow) % 2
			}
		}
		for p := leng; p < 4; p++ {
			k := 0
			for q := 0; q < 16; q++ {
				pow := 1
				for m := 15; m > q; m-- {
					pow *= 2
				}
				bt[16*p+q] = (k / pow) % 2
			}
		}
	} else {
		for i := 0; i < 4; i++ {
			k := int(str[i])
			for j := 0; j < 16; j++ {
				pow := 1
				for m := 15; m > j; m-- {
					pow *= 2
				}
				bt[16*i+j] = (k / pow) % 2
			}
		}
	}
	return bt
}

// bt4ToHex 將 4 位元二進位字串轉換為十六進位字元
func bt4ToHex(binary string) string {
	switch binary {
	case "0000":
		return "0"
	case "0001":
		return "1"
	case "0010":
		return "2"
	case "0011":
		return "3"
	case "0100":
		return "4"
	case "0101":
		return "5"
	case "0110":
		return "6"
	case "0111":
		return "7"
	case "1000":
		return "8"
	case "1001":
		return "9"
	case "1010":
		return "A"
	case "1011":
		return "B"
	case "1100":
		return "C"
	case "1101":
		return "D"
	case "1110":
		return "E"
	case "1111":
		return "F"
	}
	return "" // Should not reach here
}

// hexToBt4 將十六進位字元轉換為 4 位元二進位字串
func hexToBt4(hex string) string {
	switch hex {
	case "0":
		return "0000"
	case "1":
		return "0001"
	case "2":
		return "0010"
	case "3":
		return "0011"
	case "4":
		return "0100"
	case "5":
		return "0101"
	case "6":
		return "0110"
	case "7":
		return "0111"
	case "8":
		return "1000"
	case "9":
		return "1001"
	case "A":
		return "1010"
	case "B":
		return "1011"
	case "C":
		return "1100"
	case "D":
		return "1101"
	case "E":
		return "1110"
	case "F":
		return "1111"
	}
	return "" // Should not reach here
}

// byteToString 將 64 位元陣列轉換為字串
func byteToString(byteData []int) string {
	str := ""
	for i := 0; i < 4; i++ {
		count := 0
		for j := 0; j < 16; j++ {
			pow := 1
			for m := 15; m > j; m-- {
				pow *= 2
			}
			count += byteData[16*i+j] * pow
		}
		if count != 0 {
			str += string(rune(count))
		}
	}
	return str
}

// bt64ToHex 將 64 位元陣列轉換為十六進位字串
func bt64ToHex(byteData []int) string {
	hex := ""
	for i := 0; i < 16; i++ {
		bt := ""
		for j := 0; j < 4; j++ {
			bt += strconv.Itoa(byteData[i*4+j])
		}
		hex += bt4ToHex(bt)
	}
	return hex
}

// hexToBt64 將十六進位字串轉換為 64 位元二進位字串
func hexToBt64(hex string) string {
	binary := ""
	for i := 0; i < 16; i++ {
		binary += hexToBt4(string(hex[i]))
	}
	return binary
}

// enc 函數實現了 DES 核心加密演算法
func enc(dataByte, keyByte []int) []int {
	keys := generateKeys(keyByte)
	ipByte := initPermute(dataByte)
	ipLeft := make([]int, 32)
	ipRight := make([]int, 32)
	tempLeft := make([]int, 32)

	for k := 0; k < 32; k++ {
		ipLeft[k] = ipByte[k]
		ipRight[k] = ipByte[32+k]
	}

	for i := 0; i < 16; i++ {
		for j := 0; j < 32; j++ {
			tempLeft[j] = ipLeft[j]
			ipLeft[j] = ipRight[j]
		}
		key := make([]int, 48)
		for m := 0; m < 48; m++ {
			key[m] = keys[i][m]
		}
		tempRight := xor(pPermute(sBoxPermute(xor(expandPermute(ipRight), key))), tempLeft)
		for n := 0; n < 32; n++ {
			ipRight[n] = tempRight[n]
		}
	}

	finalData := make([]int, 64)
	for i := 0; i < 32; i++ {
		finalData[i] = ipRight[i]
		finalData[32+i] = ipLeft[i]
	}
	return finallyPermute(finalData)
}

// dec 函數實現了 DES 核心解密演算法
func dec(dataByte, keyByte []int) []int {
	keys := generateKeys(keyByte)
	ipByte := initPermute(dataByte)
	ipLeft := make([]int, 32)
	ipRight := make([]int, 32)
	tempLeft := make([]int, 32)

	for k := 0; k < 32; k++ {
		ipLeft[k] = ipByte[k]
		ipRight[k] = ipByte[32+k]
	}

	for i := 15; i >= 0; i-- {
		for j := 0; j < 32; j++ {
			tempLeft[j] = ipLeft[j]
			ipLeft[j] = ipRight[j]
		}
		key := make([]int, 48)
		for m := 0; m < 48; m++ {
			key[m] = keys[i][m]
		}

		tempRight := xor(pPermute(sBoxPermute(xor(expandPermute(ipRight), key))), tempLeft)
		for n := 0; n < 32; n++ {
			ipRight[n] = tempRight[n]
		}
	}

	finalData := make([]int, 64)
	for i := 0; i < 32; i++ {
		finalData[i] = ipRight[i]
		finalData[32+i] = ipLeft[i]
	}
	return finallyPermute(finalData)
}

// initPermute 初始置換 (IP)
func initPermute(originalData []int) []int {
	ipByte := make([]int, 64)
	for i, m, n := 0, 1, 0; i < 4; i, m, n = i+1, m+2, n+2 {
		for j, k := 7, 0; j >= 0; j, k = j-1, k+1 {
			ipByte[i*8+k] = originalData[j*8+m]
			ipByte[i*8+k+32] = originalData[j*8+n]
		}
	}
	return ipByte
}

// expandPermute 擴展置換 (E-box)
func expandPermute(rightData []int) []int {
	epByte := make([]int, 48)
	for i := 0; i < 8; i++ {
		if i == 0 {
			epByte[i*6+0] = rightData[31]
		} else {
			epByte[i*6+0] = rightData[i*4-1]
		}
		epByte[i*6+1] = rightData[i*4+0]
		epByte[i*6+2] = rightData[i*4+1]
		epByte[i*6+3] = rightData[i*4+2]
		epByte[i*6+4] = rightData[i*4+3]
		if i == 7 {
			epByte[i*6+5] = rightData[0]
		} else {
			epByte[i*6+5] = rightData[i*4+4]
		}
	}
	return epByte
}

// xor 執行位元異或操作
func xor(byteOne, byteTwo []int) []int {
	xorByte := make([]int, len(byteOne))
	for i := 0; i < len(byteOne); i++ {
		xorByte[i] = byteOne[i] ^ byteTwo[i]
	}
	return xorByte
}

// sBoxPermute S-box 置換
func sBoxPermute(expandByte []int) []int {
	sBoxByte := make([]int, 32)
	binary := ""

	s1 := [][]int{
		{14, 4, 13, 1, 2, 15, 11, 8, 3, 10, 6, 12, 5, 9, 0, 7},
		{0, 15, 7, 4, 14, 2, 13, 1, 10, 6, 12, 11, 9, 5, 3, 8},
		{4, 1, 14, 8, 13, 6, 2, 11, 15, 12, 9, 7, 3, 10, 5, 0},
		{15, 12, 8, 2, 4, 9, 1, 7, 5, 11, 3, 14, 10, 0, 6, 13}}

	s2 := [][]int{
		{15, 1, 8, 14, 6, 11, 3, 4, 9, 7, 2, 13, 12, 0, 5, 10},
		{3, 13, 4, 7, 15, 2, 8, 14, 12, 0, 1, 10, 6, 9, 11, 5},
		{0, 14, 7, 11, 10, 4, 13, 1, 5, 8, 12, 6, 9, 3, 2, 15},
		{13, 8, 10, 1, 3, 15, 4, 2, 11, 6, 7, 12, 0, 5, 14, 9}}

	s3 := [][]int{
		{10, 0, 9, 14, 6, 3, 15, 5, 1, 13, 12, 7, 11, 4, 2, 8},
		{13, 7, 0, 9, 3, 4, 6, 10, 2, 8, 5, 14, 12, 11, 15, 1},
		{13, 6, 4, 9, 8, 15, 3, 0, 11, 1, 2, 12, 5, 10, 14, 7},
		{1, 10, 13, 0, 6, 9, 8, 7, 4, 15, 14, 3, 11, 5, 2, 12}}

	s4 := [][]int{
		{7, 13, 14, 3, 0, 6, 9, 10, 1, 2, 8, 5, 11, 12, 4, 15},
		{13, 8, 11, 5, 6, 15, 0, 3, 4, 7, 2, 12, 1, 10, 14, 9},
		{10, 6, 9, 0, 12, 11, 7, 13, 15, 1, 3, 14, 5, 2, 8, 4},
		{3, 15, 0, 6, 10, 1, 13, 8, 9, 4, 5, 11, 12, 7, 2, 14}}

	s5 := [][]int{
		{2, 12, 4, 1, 7, 10, 11, 6, 8, 5, 3, 15, 13, 0, 14, 9},
		{14, 11, 2, 12, 4, 7, 13, 1, 5, 0, 15, 10, 3, 9, 8, 6},
		{4, 2, 1, 11, 10, 13, 7, 8, 15, 9, 12, 5, 6, 3, 0, 14},
		{11, 8, 12, 7, 1, 14, 2, 13, 6, 15, 0, 9, 10, 4, 5, 3}}

	s6 := [][]int{
		{12, 1, 10, 15, 9, 2, 6, 8, 0, 13, 3, 4, 14, 7, 5, 11},
		{10, 15, 4, 2, 7, 12, 9, 5, 6, 1, 13, 14, 0, 11, 3, 8},
		{9, 14, 15, 5, 2, 8, 12, 3, 7, 0, 4, 10, 1, 13, 11, 6},
		{4, 3, 2, 12, 9, 5, 15, 10, 11, 14, 1, 7, 6, 0, 8, 13}}

	s7 := [][]int{
		{4, 11, 2, 14, 15, 0, 8, 13, 3, 12, 9, 7, 5, 10, 6, 1},
		{13, 0, 11, 7, 4, 9, 1, 10, 14, 3, 5, 12, 2, 15, 8, 6},
		{1, 4, 11, 13, 12, 3, 7, 14, 10, 15, 6, 8, 0, 5, 9, 2},
		{6, 11, 13, 8, 1, 4, 10, 7, 9, 5, 0, 15, 14, 2, 3, 12}}

	s8 := [][]int{
		{13, 2, 8, 4, 6, 15, 11, 1, 10, 9, 3, 14, 5, 0, 12, 7},
		{1, 15, 13, 8, 10, 3, 7, 4, 12, 5, 6, 11, 0, 14, 9, 2},
		{7, 11, 4, 1, 9, 12, 14, 2, 0, 6, 10, 13, 15, 3, 5, 8},
		{2, 1, 14, 7, 4, 10, 8, 13, 15, 12, 9, 0, 3, 5, 6, 11}}

	sBoxes := [][][]int{s1, s2, s3, s4, s5, s6, s7, s8}

	for m := 0; m < 8; m++ {
		i := expandByte[m*6+0]*2 + expandByte[m*6+5]
		j := expandByte[m*6+1]*8 + expandByte[m*6+2]*4 + expandByte[m*6+3]*2 + expandByte[m*6+4]

		binary = getBoxBinary(sBoxes[m][i][j])

		sBoxByte[m*4+0] = int(binary[0] - '0')
		sBoxByte[m*4+1] = int(binary[1] - '0')
		sBoxByte[m*4+2] = int(binary[2] - '0')
		sBoxByte[m*4+3] = int(binary[3] - '0')
	}
	return sBoxByte
}

// pPermute P-box 置換
func pPermute(sBoxByte []int) []int {
	pBoxPermute := make([]int, 32)
	pBoxPermute[0] = sBoxByte[15]
	pBoxPermute[1] = sBoxByte[6]
	pBoxPermute[2] = sBoxByte[19]
	pBoxPermute[3] = sBoxByte[20]
	pBoxPermute[4] = sBoxByte[28]
	pBoxPermute[5] = sBoxByte[11]
	pBoxPermute[6] = sBoxByte[27]
	pBoxPermute[7] = sBoxByte[16]
	pBoxPermute[8] = sBoxByte[0]
	pBoxPermute[9] = sBoxByte[14]
	pBoxPermute[10] = sBoxByte[22]
	pBoxPermute[11] = sBoxByte[25]
	pBoxPermute[12] = sBoxByte[4]
	pBoxPermute[13] = sBoxByte[17]
	pBoxPermute[14] = sBoxByte[30]
	pBoxPermute[15] = sBoxByte[9]
	pBoxPermute[16] = sBoxByte[1]
	pBoxPermute[17] = sBoxByte[7]
	pBoxPermute[18] = sBoxByte[23]
	pBoxPermute[19] = sBoxByte[13]
	pBoxPermute[20] = sBoxByte[31]
	pBoxPermute[21] = sBoxByte[26]
	pBoxPermute[22] = sBoxByte[2]
	pBoxPermute[23] = sBoxByte[8]
	pBoxPermute[24] = sBoxByte[18]
	pBoxPermute[25] = sBoxByte[12]
	pBoxPermute[26] = sBoxByte[29]
	pBoxPermute[27] = sBoxByte[5]
	pBoxPermute[28] = sBoxByte[21]
	pBoxPermute[29] = sBoxByte[10]
	pBoxPermute[30] = sBoxByte[3]
	pBoxPermute[31] = sBoxByte[24]
	return pBoxPermute
}

// finallyPermute 最終置換 (FP)
func finallyPermute(endByte []int) []int {
	fpByte := make([]int, 64)
	fpByte[0] = endByte[39]
	fpByte[1] = endByte[7]
	fpByte[2] = endByte[47]
	fpByte[3] = endByte[15]
	fpByte[4] = endByte[55]
	fpByte[5] = endByte[23]
	fpByte[6] = endByte[63]
	fpByte[7] = endByte[31]
	fpByte[8] = endByte[38]
	fpByte[9] = endByte[6]
	fpByte[10] = endByte[46]
	fpByte[11] = endByte[14]
	fpByte[12] = endByte[54]
	fpByte[13] = endByte[22]
	fpByte[14] = endByte[62]
	fpByte[15] = endByte[30]
	fpByte[16] = endByte[37]
	fpByte[17] = endByte[5]
	fpByte[18] = endByte[45]
	fpByte[19] = endByte[13]
	fpByte[20] = endByte[53]
	fpByte[21] = endByte[21]
	fpByte[22] = endByte[61]
	fpByte[23] = endByte[29]
	fpByte[24] = endByte[36]
	fpByte[25] = endByte[4]
	fpByte[26] = endByte[44]
	fpByte[27] = endByte[12]
	fpByte[28] = endByte[52]
	fpByte[29] = endByte[20]
	fpByte[30] = endByte[60]
	fpByte[31] = endByte[28]
	fpByte[32] = endByte[35]
	fpByte[33] = endByte[3]
	fpByte[34] = endByte[43]
	fpByte[35] = endByte[11]
	fpByte[36] = endByte[51]
	fpByte[37] = endByte[19]
	fpByte[38] = endByte[59]
	fpByte[39] = endByte[27]
	fpByte[40] = endByte[34]
	fpByte[41] = endByte[2]
	fpByte[42] = endByte[42]
	fpByte[43] = endByte[10]
	fpByte[44] = endByte[50]
	fpByte[45] = endByte[18]
	fpByte[46] = endByte[58]
	fpByte[47] = endByte[26]
	fpByte[48] = endByte[33]
	fpByte[49] = endByte[1]
	fpByte[50] = endByte[41]
	fpByte[51] = endByte[9]
	fpByte[52] = endByte[49]
	fpByte[53] = endByte[17]
	fpByte[54] = endByte[57]
	fpByte[55] = endByte[25]
	fpByte[56] = endByte[32]
	fpByte[57] = endByte[0]
	fpByte[58] = endByte[40]
	fpByte[59] = endByte[8]
	fpByte[60] = endByte[48]
	fpByte[61] = endByte[16]
	fpByte[62] = endByte[56]
	fpByte[63] = endByte[24]
	return fpByte
}

// getBoxBinary 將整數轉換為 4 位元二進位字串
func getBoxBinary(i int) string {
	return fmt.Sprintf("%04b", i)
}

// generateKeys 生成 16 個子金鑰
func generateKeys(keyByte []int) [][]int {
	key := make([]int, 56)
	keys := make([][]int, 16)
	for i := range keys {
		keys[i] = make([]int, 48)
	}

	loop := []int{1, 1, 2, 2, 2, 2, 2, 2, 1, 2, 2, 2, 2, 2, 2, 1}

	// 複製原始 JavaScript 程式碼中非標準的金鑰初始置換邏輯
	// Original JavaScript logic:
	// for (i = 0; i < 7; i++) {
	//     for (j = 0, k = 7; j < 8; j++, k--) {
	//         key[i * 8 + j] = keyByte[8 * k + i];
	//     }
	// }
	for i := 0; i < 7; i++ {
		for j, k := 0, 7; j < 8; j, k = j+1, k-1 {
			key[i*8+j] = keyByte[8*k+i]
		}
	}

	for i := 0; i < 16; i++ {
		// 左移操作
		tempLeft := make([]int, 28)
		tempRight := make([]int, 28)
		copy(tempLeft, key[0:28])
		copy(tempRight, key[28:56])

		for j := 0; j < loop[i]; j++ {
			// 循環左移
			firstLeftBit := tempLeft[0]
			for k := 0; k < 27; k++ {
				tempLeft[k] = tempLeft[k+1]
			}
			tempLeft[27] = firstLeftBit

			firstRightBit := tempRight[0]
			for k := 0; k < 27; k++ {
				tempRight[k] = tempRight[k+1]
			}
			tempRight[27] = firstRightBit
		}

		copy(key[0:28], tempLeft)
		copy(key[28:56], tempRight)

		// 置換 PC-2
		pc2 := []int{
			14, 17, 11, 24, 1, 5, 3, 28,
			15, 6, 21, 10, 23, 19, 12, 4,
			26, 8, 16, 7, 27, 20, 13, 2,
			41, 52, 31, 37, 47, 55, 30, 40,
			51, 45, 33, 48, 44, 49, 39, 56,
			34, 53, 46, 42, 50, 36, 29, 32,
		}
		for m := 0; m < 48; m++ {
			keys[i][m] = key[pc2[m]-1]
		}
	}
	return keys
}

// func main() {
// 	// 範例用法
// 	data := "HelloGo!"
// 	key := "12345678" // 8 字元金鑰

// 	// 將金鑰轉換為位元陣列，Go 語言中的 DES 實現通常需要 64 位元 (8 位元組) 金鑰
// 	// 這裡我們需要將金鑰字串轉換為 64 個位元
// 	//keyBytes := strToBt(key)

// 	// 加密
// 	encryptedData := StrEnc(data, key, "", "") // 單 DES
// 	fmt.Printf("原始數據: %s\n", data)
// 	fmt.Printf("加密數據 (Hex): %s\n", encryptedData)

// 	// 解密
// 	decryptedData := StrDec(encryptedData, key, "", "") // 單 DES
// 	fmt.Printf("解密數據: %s\n", decryptedData)

// 	// 3DES 範例
// 	data3DES := "2023212762Ricxx0809.LT-326991-vNraF3YTmA4L7uHkeN5bOjfQkunpwn-cas"
// 	key1 := "1"
// 	key2 := "2"
// 	key3 := "3"

// 	encrypted3DES := StrEnc(data3DES, key1, key2, key3)
// 	fmt.Printf("\n原始數據 (3DES): %s\n", data3DES)
// 	fmt.Printf("加密數據 (3DES Hex): %s\n", encrypted3DES)

// 	decrypted3DES := StrDec(encrypted3DES, key1, key2, key3)
// 	fmt.Printf("解密數據 (3DES): %s\n", decrypted3DES)
// }
