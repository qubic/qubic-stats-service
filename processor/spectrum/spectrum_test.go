package spectrum

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_RankRichList_thenOrderedByBalanceDescending(t *testing.T) {

	richList := RichList{
		{Identity: "B", Balance: 10},
		{Identity: "A", Balance: 300},
		{Identity: "C", Balance: 200},
	}

	rankRichList(richList)

	require.Len(t, richList, 3)
	assert.Equal(t, RichList{
		{Rank: 0, Identity: "A", Balance: 300},
		{Rank: 1, Identity: "C", Balance: 200},
		{Rank: 2, Identity: "B", Balance: 10},
	}, richList)
}

func Test_RankRichList_givenEqualBalances_thenRankIsStable(t *testing.T) {

	first := RichList{
		{Identity: "C", Balance: 100},
		{Identity: "A", Balance: 100},
		{Identity: "B", Balance: 100},
	}
	second := RichList{
		{Identity: "B", Balance: 100},
		{Identity: "C", Balance: 100},
		{Identity: "A", Balance: 100},
	}

	rankRichList(first)
	rankRichList(second)

	assert.Equal(t, first, second)
	assert.Equal(t, "A", first[0].Identity) // the identity breaks the tie
}

func Test_RankRichList_givenEmptyList_thenNothingHappens(t *testing.T) {

	var richList RichList

	rankRichList(richList)

	assert.Empty(t, richList)
}
