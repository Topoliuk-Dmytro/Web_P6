package calc

import (
	"errors"
	"math"
)

// Equipment describes one group of identical electric receivers (ЕП).
type Equipment struct {
	Name   string
	Eta    float64 // ηн
	CosPhi float64 // cos φ
	UN     float64 // Uн, кВ
	Count  int     // n, шт
	Pn     float64 // Pн, кВт
	Kv     float64 // Kв
	TgPhi  float64 // tg φ
}

// VariantParams stores parameters that differ for each variant (table 6.8).
type VariantParams struct {
	GrinderPn      float64 // Рн for grinding machine
	PolisherKv     float64 // Kв for polishing machine
	CircularSawTg  float64 // tg φ for circular saw
}

// Result contains the main outputs needed for the report (items 1.1–1.14).
type Result struct {
	Variant int

	// For one shop (ШР1 = ШР2 = ШР3).
	ShopKv       float64
	ShopNe       float64
	ShopKp       float64
	ShopPp       float64
	ShopQp       float64
	ShopSp       float64
	ShopIp       float64

	// For the whole workshop (весь цех).
	TotalKv      float64
	TotalNe      float64
	TotalKp      float64
	TotalPp      float64
	TotalQp      float64
	TotalSp      float64
	TotalIp      float64
}

// Hard‑coded coefficients from the method (for this task).
// For level I (Т0 = 10 хв) and the given group of loads.
const shopKp = 1.25

// For the whole workshop (Т0 = 2,5 год).
const totalKp = 0.7

// variantTable encodes table 6.8 (parameters that change with variant).
var variantTable = map[int]VariantParams{
	1: {GrinderPn: 20, PolisherKv: 0.21, CircularSawTg: 1.55},
	2: {GrinderPn: 21, PolisherKv: 0.22, CircularSawTg: 1.56},
	3: {GrinderPn: 22, PolisherKv: 0.23, CircularSawTg: 1.57},
	4: {GrinderPn: 23, PolisherKv: 0.24, CircularSawTg: 1.58},
	5: {GrinderPn: 24, PolisherKv: 0.25, CircularSawTg: 1.59},
	6: {GrinderPn: 25, PolisherKv: 0.26, CircularSawTg: 1.61},
	7: {GrinderPn: 26, PolisherKv: 0.27, CircularSawTg: 1.62},
	8: {GrinderPn: 27, PolisherKv: 0.28, CircularSawTg: 1.63},
	9: {GrinderPn: 28, PolisherKv: 0.29, CircularSawTg: 1.64},
	0: {GrinderPn: 29, PolisherKv: 0.31, CircularSawTg: 1.65},
}

// baseShopEquipment returns the list of equipment for one shop (ШР1)
// for the control example (table 6.7), without variant adjustments.
func baseShopEquipment() []Equipment {
	return []Equipment{
		{
			Name:   "Шліфувальний верстат (1-4)",
			Eta:    0.92,
			CosPhi: 0.9,
			UN:     0.38,
			Count:  4,
			Pn:     20,
			Kv:     0.15,
			TgPhi:  1.33,
		},
		{
			Name:   "Свердлильний верстат (5-6)",
			Eta:    0.92,
			CosPhi: 0.9,
			UN:     0.38,
			Count:  2,
			Pn:     14,
			Kv:     0.12,
			TgPhi:  1.0,
		},
		{
			Name:   "Фугувальний верстат (9-12)",
			Eta:    0.92,
			CosPhi: 0.9,
			UN:     0.38,
			Count:  4,
			Pn:     42,
			Kv:     0.15,
			TgPhi:  1.33,
		},
		{
			Name:   "Циркулярна пила (13)",
			Eta:    0.92,
			CosPhi: 0.9,
			UN:     0.38,
			Count:  1,
			Pn:     36,
			Kv:     0.3,
			TgPhi:  1.52,
		},
		{
			Name:   "Прес (16)",
			Eta:    0.92,
			CosPhi: 0.9,
			UN:     0.38,
			Count:  1,
			Pn:     20,
			Kv:     0.5,
			TgPhi:  0.75,
		},
		{
			Name:   "Полірувальний верстат (24)",
			Eta:    0.92,
			CosPhi: 0.9,
			UN:     0.38,
			Count:  1,
			Pn:     40,
			Kv:     0.2,
			TgPhi:  1.0,
		},
		{
			Name:   "Фрезерний верстат (26-27)",
			Eta:    0.92,
			CosPhi: 0.9,
			UN:     0.38,
			Count:  2,
			Pn:     32,
			Kv:     0.2,
			TgPhi:  1.0,
		},
		{
			Name:   "Вентилятор (36)",
			Eta:    0.92,
			CosPhi: 0.9,
			UN:     0.38,
			Count:  1,
			Pn:     20,
			Kv:     0.65,
			TgPhi:  0.75,
		},
	}
}

// bigConsumers returns equipment that is powered directly from the substation (ТП).
func bigConsumers() []Equipment {
	return []Equipment{
		{
			Name:   "Зварювальний трансформатор",
			Eta:    0.92,
			CosPhi: 0.9,
			UN:     0.38,
			Count:  2,
			Pn:     100,
			Kv:     0.2,
			TgPhi:  3.0,
		},
		{
			Name:   "Сушильна шафа",
			Eta:    0.92,
			CosPhi: 0.9,
			UN:     0.38,
			Count:  2,
			Pn:     120,
			Kv:     0.8,
			TgPhi:  0.0, // tg φ не задано в таблиці, вважаємо приблизно 0
		},
	}
}

// ApplyVariant applies adjustments from table 6.8 to the base equipment of one shop.
func ApplyVariant(eqs []Equipment, v int) []Equipment {
	params, ok := variantTable[v]
	if !ok {
		// If variant is out of range, fall back to 1.
		params = variantTable[1]
	}

	out := make([]Equipment, len(eqs))
	copy(out, eqs)

	for i := range out {
		switch {
		case out[i].Name == "Шліфувальний верстат (1-4)":
			out[i].Pn = params.GrinderPn
		case out[i].Name == "Полірувальний верстат (24)":
			out[i].Kv = params.PolisherKv
		case out[i].Name == "Циркулярна пила (13)":
			out[i].TgPhi = params.CircularSawTg
		}
	}

	return out
}

// Calculate performs all necessary calculations for a given variant (0–9).
func Calculate(variant int) (*Result, error) {
	if variant < 0 || variant > 9 {
		return nil, errors.New("варіант має бути в діапазоні 0–9")
	}

	shopEq := ApplyVariant(baseShopEquipment(), variant)
	bigEq := bigConsumers()

	// One shop aggregates.
	shopPsum, shopPkvSum, shopPkvTgSum, shopP2Sum := aggregate(shopEq)
	shopKv := 0.0
	shopNe := 0.0
	if shopPsum > 0 {
		shopKv = shopPkvSum / shopPsum
		shopNe = (shopPsum * shopPsum) / shopP2Sum
	}

	shopPp := shopKp * shopPkvSum
	shopQp := 1.0 * shopPkvTgSum
	shopSp := math.Hypot(shopPp, shopQp)
	shopIp := currentFromPower(shopSp, 0.38)

	// Whole workshop (three identical shops + big consumers).
	totalPsum := 3*shopPsum + sumP(baseShopEquipment(), bigEq)
	totalPkvSum := 3*shopPkvSum + sumPkv(baseShopEquipment(), bigEq)
	totalPkvTgSum := 3*shopPkvTgSum + sumPkvTg(baseShopEquipment(), bigEq)
	totalP2Sum := 3*shopP2Sum + sumP2(baseShopEquipment(), bigEq)

	totalKv := 0.0
	totalNe := 0.0
	if totalPsum > 0 {
		totalKv = totalPkvSum / totalPsum
		totalNe = (totalPsum * totalPsum) / totalP2Sum
	}

	totalPp := totalKp * totalPkvSum
	totalQp := totalKp * totalPkvTgSum
	totalSp := math.Hypot(totalPp, totalQp)
	totalIp := currentFromPower(totalSp, 0.38)

	return &Result{
		Variant: variant,

		ShopKv: shopKv,
		ShopNe: shopNe,
		ShopKp: shopKp,
		ShopPp: shopPp,
		ShopQp: shopQp,
		ShopSp: shopSp,
		ShopIp: shopIp,

		TotalKv: totalKv,
		TotalNe: totalNe,
		TotalKp: totalKp,
		TotalPp: totalPp,
		TotalQp: totalQp,
		TotalSp: totalSp,
		TotalIp: totalIp,
	}, nil
}

// aggregate computes basic sums for a set of equipment.
func aggregate(list []Equipment) (sumP, sumPkv, sumPkvTg, sumP2 float64) {
	for _, e := range list {
		p := float64(e.Count) * e.Pn
		sumP += p
		sumPkv += p * e.Kv
		sumPkvTg += p * e.Kv * e.TgPhi
		sumP2 += p * p
	}
	return
}

// Helper sums including both shop equipment and large consumers.
func sumP(shop []Equipment, big []Equipment) float64 {
	p, _, _, _ := aggregate(append([]Equipment{}, append(shop, big...)...))
	return p
}

func sumPkv(shop []Equipment, big []Equipment) float64 {
	_, pkv, _, _ := aggregate(append([]Equipment{}, append(shop, big...)...))
	return pkv
}

func sumPkvTg(shop []Equipment, big []Equipment) float64 {
	_, _, pkvTg, _ := aggregate(append([]Equipment{}, append(shop, big...)...))
	return pkvTg
}

func sumP2(shop []Equipment, big []Equipment) float64 {
	_, _, _, p2 := aggregate(append([]Equipment{}, append(shop, big...)...))
	return p2
}

// currentFromPower calculates line current (A) for three‑phase load from full power (кВА) and line voltage (кВ).
func currentFromPower(sp float64, uLine float64) float64 {
	if sp <= 0 || uLine <= 0 {
		return 0
	}
	return sp / (math.Sqrt(3) * uLine)
}

