// Copyright 2020 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package unique

import (
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
)

func TestUniquifyByteSlices(t *testing.T) {
	tests := []struct {
		input    []string
		expected []string
	}{
		{
			input:    []string{"foo", "foo"},
			expected: []string{"foo"},
		},
		{
			input:    []string{},
			expected: []string{},
		},
		{
			input:    []string{"", ""},
			expected: []string{""},
		},
		{
			input:    []string{"foo"},
			expected: []string{"foo"},
		},
		{
			input:    []string{"foo", "bar", "foo"},
			expected: []string{"bar", "foo"},
		},
		{
			input:    []string{"foo", "bar"},
			expected: []string{"bar", "foo"},
		},
		{
			input:    []string{"bar", "bar", "foo"},
			expected: []string{"bar", "foo"},
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			input := make([][]byte, len(tt.input))
			expected := make([][]byte, len(tt.expected))
			for i := range tt.input {
				input[i] = []byte(tt.input[i])
			}
			for i := range tt.expected {
				expected[i] = []byte(tt.expected[i])
			}
			if got := UniquifyByteSlices(input); !reflect.DeepEqual(got, expected) {
				t.Errorf("UniquifyByteSlices() = %v, expected %v", got, expected)
			}
		})
	}
}

func TestUniquifySpans(t *testing.T) {
	tests := []struct {
		input    [][][]string
		expected [][][]string
	}{
		{
			input:    [][][]string{{{"a", "b"}}, {{"a", "b"}}},
			expected: [][][]string{{{"a", "b"}}},
		},
		{
			input:    [][][]string{},
			expected: [][][]string{},
		},
		{
			input:    [][][]string{{{"a", "b"}}},
			expected: [][][]string{{{"a", "b"}}},
		},
		{
			input:    [][][]string{{{"a", "b"}}, {{"a", "b"}, {"c", "d"}}, {{"a", "b"}}},
			expected: [][][]string{{{"a", "b"}}, {{"a", "b"}, {"c", "d"}}},
		},
		{
			input:    [][][]string{{{"a", "b"}, {"c", "d"}}, {{"a", "b"}}},
			expected: [][][]string{{{"a", "b"}}, {{"a", "b"}, {"c", "d"}}},
		},
		{
			input:    [][][]string{{{"bar", "foo"}}, {{"bar", "foo"}}, {{"foobar", "foobaz"}}},
			expected: [][][]string{{{"bar", "foo"}}, {{"foobar", "foobaz"}}},
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			// Add a random permutation within each span set.
			for idx := range tt.input {
				rand.Shuffle(len(tt.input[idx]), func(i, j int) {
					tt.input[idx][i], tt.input[idx][j] = tt.input[idx][j], tt.input[idx][i]
				})
			}

			// Add a random permutation at the top level.
			rand.Shuffle(len(tt.input), func(i, j int) {
				tt.input[i], tt.input[j] = tt.input[j], tt.input[i]
			})

			input := make([]roachpb.Spans, len(tt.input))
			expected := make([]roachpb.Spans, len(tt.expected))
			for i, spans := range tt.input {
				input[i] = make(roachpb.Spans, len(spans))
				for j := range spans {
					input[i][j] = roachpb.Span{
						Key:    roachpb.Key(spans[j][0]),
						EndKey: roachpb.Key(spans[j][1]),
					}
				}
			}
			for i, spans := range tt.expected {
				expected[i] = make(roachpb.Spans, len(spans))
				for j := range spans {
					expected[i][j] = roachpb.Span{
						Key:    roachpb.Key(spans[j][0]),
						EndKey: roachpb.Key(spans[j][1]),
					}
				}
			}
			if got := SortAndUniquifySpanSets(input); !reflect.DeepEqual(got, expected) {
				t.Errorf("SortAndUniquifySpanSets() = %v, expected %v", got, expected)
			}
		})
	}
}

type uasTestCase = struct {
	left          []int
	right         []int
	expectedLeft  []int
	expectedRight []int
}

func TestUniquifyAcrossSlices(t *testing.T) {
	tests := []uasTestCase{
		{
			left:          []int{0, 5, 7, 10},
			right:         []int{1, 5, 7, 11},
			expectedLeft:  []int{0, 10},
			expectedRight: []int{1, 11},
		},
		{
			left:          []int{0, 5, 7, 10},
			right:         []int{},
			expectedLeft:  []int{0, 5, 7, 10},
			expectedRight: []int{},
		},
		{
			left:          []int{},
			right:         []int{},
			expectedLeft:  []int{},
			expectedRight: []int{},
		},
		{
			left:          []int{3, 5, 7},
			right:         []int{7},
			expectedLeft:  []int{3, 5},
			expectedRight: []int{},
		},
		{
			left:          []int{3, 5, 7},
			right:         []int{8},
			expectedLeft:  []int{3, 5, 7},
			expectedRight: []int{8},
		},
		{
			left:          []int{1, 2, 3},
			right:         []int{1, 2, 3},
			expectedLeft:  []int{},
			expectedRight: []int{},
		},
	}

	origTests := tests
	for _, test := range origTests {
		// For each test case, add a flipped test case.
		rightCopy := make([]int, len(test.right))
		leftCopy := make([]int, len(test.left))
		for i := range rightCopy {
			rightCopy[i] = test.right[i]
		}
		for i := range leftCopy {
			leftCopy[i] = test.left[i]
		}
		tests = append(tests, uasTestCase{
			left:          rightCopy,
			right:         leftCopy,
			expectedLeft:  test.expectedRight,
			expectedRight: test.expectedLeft,
		})
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			leftLen, rightLen := UniquifyAcrossSlices(tt.left, tt.right,
				func(l, r int) int {
					if tt.left[l] < tt.right[r] {
						return -1
					} else if tt.left[l] == tt.right[r] {
						return 0
					}
					return 1
				},
				func(i, j int) {
					tt.left[i] = tt.left[j]
				},
				func(i, j int) {
					tt.right[i] = tt.right[j]
				},
			)
			left := tt.left[:leftLen]
			right := tt.right[:rightLen]
			if !reflect.DeepEqual(left, tt.expectedLeft) {
				t.Errorf("expected %v, got %v", tt.expectedLeft, left)
			}
			if !reflect.DeepEqual(right, tt.expectedRight) {
				t.Errorf("expected %v, got %v", tt.expectedRight, right)
			}
		})
	}
}
