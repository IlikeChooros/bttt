package uttt

var _historyHeuristic = [2][9][9]int{}

func _UpdateHistory(side int, move PosType, bonus int) {
	_historyHeuristic[side][move.BigIndex()][move.SmallIndex()] += bonus

	// const (
	// 	_MaxHistoryValue = 5000
	// 	_MinHistoryValue = -_MaxHistoryValue
	// )

	// if bonus > _MaxHistoryValue {
	// 	bonus = _MaxHistoryValue
	// } else if bonus < _MinHistoryValue {
	// 	bonus = _MinHistoryValue
	// }

	// bi, si := move.BigIndex(), move.SmallIndex()

	// _historyHeuristic[side][bi][si] += int(
	// 	float64(bonus) - math.Abs(float64(bonus))*float64(_historyHeuristic[side][bi][si])/_MaxHistoryValue,
	// )
}
