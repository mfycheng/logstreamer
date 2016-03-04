// Copyright 2016 Mike Cheng
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package logstreamer

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNoHistory(t *testing.T) {
	stream := NewBufferedLogStream(10)
	obs := stream.NewObserver()

	go func() {
		defer obs.Close()
		for i := 0; i < 20; i++ {
			stream.WriteLine(fmt.Sprintf("Log %v", i))
		}
	}()

	i := 0
	for l := range obs.Chan() {
		assert.True(t, time.Now().Sub(l.Timestamp) > 0)
		assert.EqualValues(t, i, l.Number)
		assert.Equal(t, fmt.Sprintf("Log %v", i), l.Line)
		i++
	}
}

func TestHistory(t *testing.T) {
	stream := NewBufferedLogStream(10)

	// Create some history
	for i := 0; i < 20; i++ {
		stream.WriteLine(fmt.Sprintf("Log %v", i))
	}

	obs := stream.NewObserver()
	defer obs.Close()
	for i := 10; i < 20; i++ {
		entry := <-obs.Chan()

		assert.True(t, time.Now().Sub(entry.Timestamp) > 0)
		assert.EqualValues(t, i, entry.Number)
		assert.Equal(t, fmt.Sprintf("Log %v", i), entry.Line)
	}

	stream.WriteLine(fmt.Sprintf("Log %v", 20))

	entry := <-obs.Chan()
	assert.True(t, time.Now().Sub(entry.Timestamp) > 0)
	assert.EqualValues(t, 20, entry.Number)
	assert.Equal(t, fmt.Sprintf("Log %v", 20), entry.Line)
}
