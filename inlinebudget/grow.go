package inlinebudget

// The functions below are the same body at increasing lengths: Grow01 has one
// added statement, Grow40 has forty. The compiler reports a cost for each, and
// somewhere in this list it stops reporting one and says why instead.
//
//	go build -gcflags='-m -m' ./inlinebudget 2>&1 | grep -E 'Grow[0-9]+'

func Grow01(n int) int {
	n += 1
	return n
}

func Grow02(n int) int {
	n += 1
	n += 2
	return n
}

func Grow03(n int) int {
	n += 1
	n += 2
	n += 3
	return n
}

func Grow04(n int) int {
	n += 1
	n += 2
	n += 3
	n += 4
	return n
}

func Grow05(n int) int {
	n += 1
	n += 2
	n += 3
	n += 4
	n += 5
	return n
}

func Grow06(n int) int {
	n += 1
	n += 2
	n += 3
	n += 4
	n += 5
	n += 6
	return n
}

func Grow07(n int) int {
	n += 1
	n += 2
	n += 3
	n += 4
	n += 5
	n += 6
	n += 7
	return n
}

func Grow08(n int) int {
	n += 1
	n += 2
	n += 3
	n += 4
	n += 5
	n += 6
	n += 7
	n += 8
	return n
}

func Grow09(n int) int {
	n += 1
	n += 2
	n += 3
	n += 4
	n += 5
	n += 6
	n += 7
	n += 8
	n += 9
	return n
}

func Grow10(n int) int {
	n += 1
	n += 2
	n += 3
	n += 4
	n += 5
	n += 6
	n += 7
	n += 8
	n += 9
	n += 10
	return n
}

func Grow11(n int) int {
	n += 1
	n += 2
	n += 3
	n += 4
	n += 5
	n += 6
	n += 7
	n += 8
	n += 9
	n += 10
	n += 11
	return n
}

func Grow12(n int) int {
	n += 1
	n += 2
	n += 3
	n += 4
	n += 5
	n += 6
	n += 7
	n += 8
	n += 9
	n += 10
	n += 11
	n += 12
	return n
}

func Grow13(n int) int {
	n += 1
	n += 2
	n += 3
	n += 4
	n += 5
	n += 6
	n += 7
	n += 8
	n += 9
	n += 10
	n += 11
	n += 12
	n += 13
	return n
}

func Grow14(n int) int {
	n += 1
	n += 2
	n += 3
	n += 4
	n += 5
	n += 6
	n += 7
	n += 8
	n += 9
	n += 10
	n += 11
	n += 12
	n += 13
	n += 14
	return n
}

func Grow15(n int) int {
	n += 1
	n += 2
	n += 3
	n += 4
	n += 5
	n += 6
	n += 7
	n += 8
	n += 9
	n += 10
	n += 11
	n += 12
	n += 13
	n += 14
	n += 15
	return n
}

func Grow16(n int) int {
	n += 1
	n += 2
	n += 3
	n += 4
	n += 5
	n += 6
	n += 7
	n += 8
	n += 9
	n += 10
	n += 11
	n += 12
	n += 13
	n += 14
	n += 15
	n += 16
	return n
}

func Grow17(n int) int {
	n += 1
	n += 2
	n += 3
	n += 4
	n += 5
	n += 6
	n += 7
	n += 8
	n += 9
	n += 10
	n += 11
	n += 12
	n += 13
	n += 14
	n += 15
	n += 16
	n += 17
	return n
}

func Grow18(n int) int {
	n += 1
	n += 2
	n += 3
	n += 4
	n += 5
	n += 6
	n += 7
	n += 8
	n += 9
	n += 10
	n += 11
	n += 12
	n += 13
	n += 14
	n += 15
	n += 16
	n += 17
	n += 18
	return n
}

func Grow19(n int) int {
	n += 1
	n += 2
	n += 3
	n += 4
	n += 5
	n += 6
	n += 7
	n += 8
	n += 9
	n += 10
	n += 11
	n += 12
	n += 13
	n += 14
	n += 15
	n += 16
	n += 17
	n += 18
	n += 19
	return n
}

func Grow20(n int) int {
	n += 1
	n += 2
	n += 3
	n += 4
	n += 5
	n += 6
	n += 7
	n += 8
	n += 9
	n += 10
	n += 11
	n += 12
	n += 13
	n += 14
	n += 15
	n += 16
	n += 17
	n += 18
	n += 19
	n += 20
	return n
}

func Grow21(n int) int {
	n += 1
	n += 2
	n += 3
	n += 4
	n += 5
	n += 6
	n += 7
	n += 8
	n += 9
	n += 10
	n += 11
	n += 12
	n += 13
	n += 14
	n += 15
	n += 16
	n += 17
	n += 18
	n += 19
	n += 20
	n += 21
	return n
}

func Grow22(n int) int {
	n += 1
	n += 2
	n += 3
	n += 4
	n += 5
	n += 6
	n += 7
	n += 8
	n += 9
	n += 10
	n += 11
	n += 12
	n += 13
	n += 14
	n += 15
	n += 16
	n += 17
	n += 18
	n += 19
	n += 20
	n += 21
	n += 22
	return n
}

func Grow23(n int) int {
	n += 1
	n += 2
	n += 3
	n += 4
	n += 5
	n += 6
	n += 7
	n += 8
	n += 9
	n += 10
	n += 11
	n += 12
	n += 13
	n += 14
	n += 15
	n += 16
	n += 17
	n += 18
	n += 19
	n += 20
	n += 21
	n += 22
	n += 23
	return n
}

func Grow24(n int) int {
	n += 1
	n += 2
	n += 3
	n += 4
	n += 5
	n += 6
	n += 7
	n += 8
	n += 9
	n += 10
	n += 11
	n += 12
	n += 13
	n += 14
	n += 15
	n += 16
	n += 17
	n += 18
	n += 19
	n += 20
	n += 21
	n += 22
	n += 23
	n += 24
	return n
}

func Grow25(n int) int {
	n += 1
	n += 2
	n += 3
	n += 4
	n += 5
	n += 6
	n += 7
	n += 8
	n += 9
	n += 10
	n += 11
	n += 12
	n += 13
	n += 14
	n += 15
	n += 16
	n += 17
	n += 18
	n += 19
	n += 20
	n += 21
	n += 22
	n += 23
	n += 24
	n += 25
	return n
}

func Grow26(n int) int {
	n += 1
	n += 2
	n += 3
	n += 4
	n += 5
	n += 6
	n += 7
	n += 8
	n += 9
	n += 10
	n += 11
	n += 12
	n += 13
	n += 14
	n += 15
	n += 16
	n += 17
	n += 18
	n += 19
	n += 20
	n += 21
	n += 22
	n += 23
	n += 24
	n += 25
	n += 26
	return n
}

func Grow27(n int) int {
	n += 1
	n += 2
	n += 3
	n += 4
	n += 5
	n += 6
	n += 7
	n += 8
	n += 9
	n += 10
	n += 11
	n += 12
	n += 13
	n += 14
	n += 15
	n += 16
	n += 17
	n += 18
	n += 19
	n += 20
	n += 21
	n += 22
	n += 23
	n += 24
	n += 25
	n += 26
	n += 27
	return n
}

func Grow28(n int) int {
	n += 1
	n += 2
	n += 3
	n += 4
	n += 5
	n += 6
	n += 7
	n += 8
	n += 9
	n += 10
	n += 11
	n += 12
	n += 13
	n += 14
	n += 15
	n += 16
	n += 17
	n += 18
	n += 19
	n += 20
	n += 21
	n += 22
	n += 23
	n += 24
	n += 25
	n += 26
	n += 27
	n += 28
	return n
}

func Grow29(n int) int {
	n += 1
	n += 2
	n += 3
	n += 4
	n += 5
	n += 6
	n += 7
	n += 8
	n += 9
	n += 10
	n += 11
	n += 12
	n += 13
	n += 14
	n += 15
	n += 16
	n += 17
	n += 18
	n += 19
	n += 20
	n += 21
	n += 22
	n += 23
	n += 24
	n += 25
	n += 26
	n += 27
	n += 28
	n += 29
	return n
}

func Grow30(n int) int {
	n += 1
	n += 2
	n += 3
	n += 4
	n += 5
	n += 6
	n += 7
	n += 8
	n += 9
	n += 10
	n += 11
	n += 12
	n += 13
	n += 14
	n += 15
	n += 16
	n += 17
	n += 18
	n += 19
	n += 20
	n += 21
	n += 22
	n += 23
	n += 24
	n += 25
	n += 26
	n += 27
	n += 28
	n += 29
	n += 30
	return n
}

func Grow31(n int) int {
	n += 1
	n += 2
	n += 3
	n += 4
	n += 5
	n += 6
	n += 7
	n += 8
	n += 9
	n += 10
	n += 11
	n += 12
	n += 13
	n += 14
	n += 15
	n += 16
	n += 17
	n += 18
	n += 19
	n += 20
	n += 21
	n += 22
	n += 23
	n += 24
	n += 25
	n += 26
	n += 27
	n += 28
	n += 29
	n += 30
	n += 31
	return n
}

func Grow32(n int) int {
	n += 1
	n += 2
	n += 3
	n += 4
	n += 5
	n += 6
	n += 7
	n += 8
	n += 9
	n += 10
	n += 11
	n += 12
	n += 13
	n += 14
	n += 15
	n += 16
	n += 17
	n += 18
	n += 19
	n += 20
	n += 21
	n += 22
	n += 23
	n += 24
	n += 25
	n += 26
	n += 27
	n += 28
	n += 29
	n += 30
	n += 31
	n += 32
	return n
}

func Grow33(n int) int {
	n += 1
	n += 2
	n += 3
	n += 4
	n += 5
	n += 6
	n += 7
	n += 8
	n += 9
	n += 10
	n += 11
	n += 12
	n += 13
	n += 14
	n += 15
	n += 16
	n += 17
	n += 18
	n += 19
	n += 20
	n += 21
	n += 22
	n += 23
	n += 24
	n += 25
	n += 26
	n += 27
	n += 28
	n += 29
	n += 30
	n += 31
	n += 32
	n += 33
	return n
}

func Grow34(n int) int {
	n += 1
	n += 2
	n += 3
	n += 4
	n += 5
	n += 6
	n += 7
	n += 8
	n += 9
	n += 10
	n += 11
	n += 12
	n += 13
	n += 14
	n += 15
	n += 16
	n += 17
	n += 18
	n += 19
	n += 20
	n += 21
	n += 22
	n += 23
	n += 24
	n += 25
	n += 26
	n += 27
	n += 28
	n += 29
	n += 30
	n += 31
	n += 32
	n += 33
	n += 34
	return n
}

func Grow35(n int) int {
	n += 1
	n += 2
	n += 3
	n += 4
	n += 5
	n += 6
	n += 7
	n += 8
	n += 9
	n += 10
	n += 11
	n += 12
	n += 13
	n += 14
	n += 15
	n += 16
	n += 17
	n += 18
	n += 19
	n += 20
	n += 21
	n += 22
	n += 23
	n += 24
	n += 25
	n += 26
	n += 27
	n += 28
	n += 29
	n += 30
	n += 31
	n += 32
	n += 33
	n += 34
	n += 35
	return n
}

func Grow36(n int) int {
	n += 1
	n += 2
	n += 3
	n += 4
	n += 5
	n += 6
	n += 7
	n += 8
	n += 9
	n += 10
	n += 11
	n += 12
	n += 13
	n += 14
	n += 15
	n += 16
	n += 17
	n += 18
	n += 19
	n += 20
	n += 21
	n += 22
	n += 23
	n += 24
	n += 25
	n += 26
	n += 27
	n += 28
	n += 29
	n += 30
	n += 31
	n += 32
	n += 33
	n += 34
	n += 35
	n += 36
	return n
}

func Grow37(n int) int {
	n += 1
	n += 2
	n += 3
	n += 4
	n += 5
	n += 6
	n += 7
	n += 8
	n += 9
	n += 10
	n += 11
	n += 12
	n += 13
	n += 14
	n += 15
	n += 16
	n += 17
	n += 18
	n += 19
	n += 20
	n += 21
	n += 22
	n += 23
	n += 24
	n += 25
	n += 26
	n += 27
	n += 28
	n += 29
	n += 30
	n += 31
	n += 32
	n += 33
	n += 34
	n += 35
	n += 36
	n += 37
	return n
}

func Grow38(n int) int {
	n += 1
	n += 2
	n += 3
	n += 4
	n += 5
	n += 6
	n += 7
	n += 8
	n += 9
	n += 10
	n += 11
	n += 12
	n += 13
	n += 14
	n += 15
	n += 16
	n += 17
	n += 18
	n += 19
	n += 20
	n += 21
	n += 22
	n += 23
	n += 24
	n += 25
	n += 26
	n += 27
	n += 28
	n += 29
	n += 30
	n += 31
	n += 32
	n += 33
	n += 34
	n += 35
	n += 36
	n += 37
	n += 38
	return n
}

func Grow39(n int) int {
	n += 1
	n += 2
	n += 3
	n += 4
	n += 5
	n += 6
	n += 7
	n += 8
	n += 9
	n += 10
	n += 11
	n += 12
	n += 13
	n += 14
	n += 15
	n += 16
	n += 17
	n += 18
	n += 19
	n += 20
	n += 21
	n += 22
	n += 23
	n += 24
	n += 25
	n += 26
	n += 27
	n += 28
	n += 29
	n += 30
	n += 31
	n += 32
	n += 33
	n += 34
	n += 35
	n += 36
	n += 37
	n += 38
	n += 39
	return n
}

func Grow40(n int) int {
	n += 1
	n += 2
	n += 3
	n += 4
	n += 5
	n += 6
	n += 7
	n += 8
	n += 9
	n += 10
	n += 11
	n += 12
	n += 13
	n += 14
	n += 15
	n += 16
	n += 17
	n += 18
	n += 19
	n += 20
	n += 21
	n += 22
	n += 23
	n += 24
	n += 25
	n += 26
	n += 27
	n += 28
	n += 29
	n += 30
	n += 31
	n += 32
	n += 33
	n += 34
	n += 35
	n += 36
	n += 37
	n += 38
	n += 39
	n += 40
	return n
}
