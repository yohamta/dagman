// Copyright (C) 2024 The Daguflow/Dagu Authors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package grep

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGrep(t *testing.T) {
	t.Parallel()
	wd, _ := os.Getwd()
	dir := filepath.Join(wd, "/testdata")
	for _, tc := range []struct {
		Name    string
		File    string
		Pattern string
		Opts    *Options
		Want    []*Match
		IsErr   bool
	}{
		{
			Name:    "simple",
			File:    filepath.Join(dir, "test.txt"),
			Pattern: "b",
			Want: []*Match{
				{
					LineNumber: 2,
					StartLine:  2,
					Line:       "bb",
				}},
		},
		{
			Name:    "regexp",
			File:    filepath.Join(dir, "test.txt"),
			Pattern: "^b.",
			Opts: &Options{
				IsRegexp: true,
			},
			Want: []*Match{
				{
					LineNumber: 2,
					StartLine:  2,
					Line:       "bb",
				}},
		},
		{
			Name:    "before",
			File:    filepath.Join(dir, "test.txt"),
			Pattern: "b",
			Opts: &Options{
				Before: 1,
			},
			Want: []*Match{
				{
					LineNumber: 2,
					StartLine:  1,
					Line:       "aa\nbb",
				}},
		},
		{
			Name:    "before+after",
			File:    filepath.Join(dir, "test.txt"),
			Pattern: "cc",
			Opts: &Options{
				Before: 2,
				After:  2,
			},
			Want: []*Match{
				{
					LineNumber: 3,
					StartLine:  1,
					Line:       "aa\nbb\ncc\ndd\nee",
				}},
		},
		{
			Name:    "before+after,firstline",
			File:    filepath.Join(dir, "test.txt"),
			Pattern: "aa",
			Opts: &Options{
				Before: 1,
				After:  1,
			},
			Want: []*Match{
				{
					LineNumber: 1,
					StartLine:  1,
					Line:       "aa\nbb",
				}},
		},
		{
			Name:    "before+after,lastline",
			File:    filepath.Join(dir, "test.txt"),
			Pattern: "ee",
			Opts: &Options{
				Before: 1,
				After:  1,
			},
			Want: []*Match{
				{
					LineNumber: 5,
					StartLine:  4,
					Line:       "dd\nee",
				}},
		},
		{
			Name:    "no match",
			File:    filepath.Join(dir, "test.txt"),
			Pattern: "no match text",
			IsErr:   true,
		},
		{
			Name:    "no file",
			File:    filepath.Join(dir, "dummy.txt"),
			Pattern: "aa",
			IsErr:   true,
		},
		{
			Name:    "no pattern",
			File:    filepath.Join(dir, "test.txt"),
			Pattern: "",
			IsErr:   true,
		},
		{
			Name:    "invalid regexp",
			File:    filepath.Join(dir, "test.txt"),
			Pattern: "(aa",
			Opts: &Options{
				IsRegexp: true,
			},
			IsErr: true,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			dat, _ := os.ReadFile(tc.File)
			ret, err := Grep(dat, tc.Pattern, tc.Opts)
			if tc.IsErr {
				require.Empty(t, ret)
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.Want, ret)
		})
	}
}
