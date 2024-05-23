// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package aggregate // import "go.opentelemetry.io/otel/sdk/metric/internal/aggregate"

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestLastValue(t *testing.T) {
	c := new(clock)
	t.Cleanup(c.Register())

	t.Run("Int64/DeltaLastValue", testDeltaLastValue[int64]())
	c.Reset()
	t.Run("Float64/DeltaLastValue", testDeltaLastValue[float64]())
	c.Reset()

	t.Run("Int64/CumulativeLastValue", testCumulativeLastValue[int64]())
	c.Reset()
	t.Run("Float64/CumulativeLastValue", testCumulativeLastValue[float64]())
	c.Reset()

	t.Run("Int64/DeltaPrecomputedLastValue", testDeltaPrecomputedLastValue[int64]())
	c.Reset()
	t.Run("Float64/DeltaPrecomputedLastValue", testDeltaPrecomputedLastValue[float64]())
	c.Reset()

	t.Run("Int64/CumulativePrecomputedLastValue", testCumulativePrecomputedLastValue[int64]())
	c.Reset()
	t.Run("Float64/CumulativePrecomputedLastValue", testCumulativePrecomputedLastValue[float64]())
}

func testDeltaLastValue[N int64 | float64]() func(*testing.T) {
	in, remove, out := Builder[N]{
		Temporality:      metricdata.DeltaTemporality,
		Filter:           attrFltr,
		AggregationLimit: 3,
	}.LastValue()
	ctx := context.Background()
	return test[N](in, remove, out, []teststep[N]{
		{
			// Empty output if nothing is measured.
			input:  []arg[N]{},
			expect: output{n: 0, agg: metricdata.Gauge[N]{}},
		}, {
			input: []arg[N]{
				{ctx, 1, alice},
				{ctx, -1, bob},
				{ctx, 1, fltrAlice},
				{ctx, 2, alice},
				{ctx, -10, bob},
			},
			remove: []arg[N]{},
			expect: output{
				n: 2,
				agg: metricdata.Gauge[N]{
					DataPoints: []metricdata.DataPoint[N]{
						{
							Attributes: fltrAlice,
							StartTime:  y2kPlus(1),
							Time:       y2kPlus(5),
							Value:      2,
						},
						{
							Attributes: fltrBob,
							StartTime:  y2kPlus(1),
							Time:       y2kPlus(6),
							Value:      -10,
						},
					},
				},
			},
		}, {
			// Everything resets, do not report old measurements.
			input:  []arg[N]{},
			remove: []arg[N]{},
			expect: output{n: 0, agg: metricdata.Gauge[N]{}},
		}, {
			input: []arg[N]{
				{ctx, 10, alice},
				{ctx, 3, bob},
			},
			remove: []arg[N]{},
			expect: output{
				n: 2,
				agg: metricdata.Gauge[N]{
					DataPoints: []metricdata.DataPoint[N]{
						{
							Attributes: fltrAlice,
							StartTime:  y2kPlus(8),
							Time:       y2kPlus(9),
							Value:      10,
						},
						{
							Attributes: fltrBob,
							StartTime:  y2kPlus(8),
							Time:       y2kPlus(10),
							Value:      3,
						},
					},
				},
			},
		}, {
			input: []arg[N]{
				{ctx, 1, alice},
				{ctx, 1, bob},
				// These will exceed cardinality limit.
				{ctx, 1, carol},
				{ctx, 1, dave},
			},
			remove: []arg[N]{},
			expect: output{
				n: 3,
				agg: metricdata.Gauge[N]{
					DataPoints: []metricdata.DataPoint[N]{
						{
							Attributes: fltrAlice,
							StartTime:  y2kPlus(11),
							Time:       y2kPlus(12),
							Value:      1,
						},
						{
							Attributes: fltrBob,
							StartTime:  y2kPlus(11),
							Time:       y2kPlus(13),
							Value:      1,
						},
						{
							Attributes: overflowSet,
							StartTime:  y2kPlus(11),
							Time:       y2kPlus(15),
							Value:      1,
						},
					},
				},
			},
		},
	})
}

func testCumulativeLastValue[N int64 | float64]() func(*testing.T) {
	in, remove, out := Builder[N]{
		Temporality:      metricdata.CumulativeTemporality,
		Filter:           attrFltr,
		AggregationLimit: 3,
	}.LastValue()
	ctx := context.Background()
	return test[N](in, remove, out, []teststep[N]{
		{
			// Empty output if nothing is measured.
			input:  []arg[N]{},
			expect: output{n: 0, agg: metricdata.Gauge[N]{}},
		}, {
			input: []arg[N]{
				{ctx, 1, alice},
				{ctx, -1, bob},
				{ctx, 1, fltrAlice},
				{ctx, 2, alice},
				{ctx, -10, bob},
			},
			expect: output{
				n: 2,
				agg: metricdata.Gauge[N]{
					DataPoints: []metricdata.DataPoint[N]{
						{
							Attributes: fltrAlice,
							StartTime:  y2kPlus(0),
							Time:       y2kPlus(4),
							Value:      2,
						},
						{
							Attributes: fltrBob,
							StartTime:  y2kPlus(0),
							Time:       y2kPlus(5),
							Value:      -10,
						},
					},
				},
			},
		}, {
			// Cumulative temporality means no resets.
			input: []arg[N]{},
			expect: output{
				n: 2,
				agg: metricdata.Gauge[N]{
					DataPoints: []metricdata.DataPoint[N]{
						{
							Attributes: fltrAlice,
							StartTime:  y2kPlus(0),
							Time:       y2kPlus(4),
							Value:      2,
						},
						{
							Attributes: fltrBob,
							StartTime:  y2kPlus(0),
							Time:       y2kPlus(5),
							Value:      -10,
						},
					},
				},
			},
		}, {
			input: []arg[N]{
				{ctx, 10, alice},
				{ctx, 3, bob},
			},
			expect: output{
				n: 2,
				agg: metricdata.Gauge[N]{
					DataPoints: []metricdata.DataPoint[N]{
						{
							Attributes: fltrAlice,
							StartTime:  y2kPlus(0),
							Time:       y2kPlus(6),
							Value:      10,
						},
						{
							Attributes: fltrBob,
							StartTime:  y2kPlus(0),
							Time:       y2kPlus(7),
							Value:      3,
						},
					},
				},
			},
		}, {
			input: []arg[N]{
				{ctx, 1, alice},
				{ctx, 1, bob},
				// These will exceed cardinality limit.
				{ctx, 1, carol},
				{ctx, 1, dave},
			},
			expect: output{
				n: 3,
				agg: metricdata.Gauge[N]{
					DataPoints: []metricdata.DataPoint[N]{
						{
							Attributes: fltrAlice,
							StartTime:  y2kPlus(0),
							Time:       y2kPlus(8),
							Value:      1,
						},
						{
							Attributes: fltrBob,
							StartTime:  y2kPlus(0),
							Time:       y2kPlus(9),
							Value:      1,
						},
						{
							Attributes: overflowSet,
							StartTime:  y2kPlus(0),
							Time:       y2kPlus(11),
							Value:      1,
						},
					},
				},
			},
		}, {
			input: []arg[N]{},
			remove: []arg[N]{
				{ctx, 1, fltrBob},
			},
			expect: output{
				n: 3,
				agg: metricdata.Gauge[N]{
					DataPoints: []metricdata.DataPoint[N]{
						{
							Attributes: fltrAlice,
							StartTime:  y2kPlus(0),
							Time:       y2kPlus(8),
							Value:      1,
						},
						{
							Attributes: fltrBob,
							StartTime:  y2kPlus(0),
							Time:       y2kPlus(9),
							NoRecordedValue: true,
						},
						{
							Attributes: overflowSet,
							StartTime:  y2kPlus(0),
							Time:       y2kPlus(11),
							Value:      1,
						},
					},
				},
			},
		}, {
			input: []arg[N]{},
			remove: []arg[N]{},
			expect: output{
				n: 2,
				agg: metricdata.Gauge[N]{
					DataPoints: []metricdata.DataPoint[N]{
						{
							Attributes: fltrAlice,
							StartTime:  y2kPlus(0),
							Time:       y2kPlus(8),
							Value:      1,
						},
						{
							Attributes: overflowSet,
							StartTime:  y2kPlus(0),
							Time:       y2kPlus(11),
							Value:      1,
						},
					},
				},
			},
		},
	})
}

func testDeltaPrecomputedLastValue[N int64 | float64]() func(*testing.T) {
	in, remove, out := Builder[N]{
		Temporality:      metricdata.DeltaTemporality,
		Filter:           attrFltr,
		AggregationLimit: 3,
	}.PrecomputedLastValue()
	ctx := context.Background()
	return test[N](in, remove, out, []teststep[N]{
		{
			// Empty output if nothing is measured.
			input:  []arg[N]{},
			expect: output{n: 0, agg: metricdata.Gauge[N]{}},
		}, {
			input: []arg[N]{
				{ctx, 1, alice},
				{ctx, -1, bob},
				{ctx, 1, fltrAlice},
				{ctx, 2, alice},
				{ctx, -10, bob},
			},
			remove: []arg[N]{},
			expect: output{
				n: 2,
				agg: metricdata.Gauge[N]{
					DataPoints: []metricdata.DataPoint[N]{
						{
							Attributes: fltrAlice,
							StartTime:  y2kPlus(1),
							Time:       y2kPlus(5),
							Value:      2,
						},
						{
							Attributes: fltrBob,
							StartTime:  y2kPlus(1),
							Time:       y2kPlus(6),
							Value:      -10,
						},
					},
				},
			},
		}, {
			// Everything resets, do not report old measurements.
			input:  []arg[N]{},
			remove: []arg[N]{},
			expect: output{n: 0, agg: metricdata.Gauge[N]{}},
		}, {
			input: []arg[N]{
				{ctx, 10, alice},
				{ctx, 3, bob},
			},
			remove: []arg[N]{},
			expect: output{
				n: 2,
				agg: metricdata.Gauge[N]{
					DataPoints: []metricdata.DataPoint[N]{
						{
							Attributes: fltrAlice,
							StartTime:  y2kPlus(8),
							Time:       y2kPlus(9),
							Value:      10,
						},
						{
							Attributes: fltrBob,
							StartTime:  y2kPlus(8),
							Time:       y2kPlus(10),
							Value:      3,
						},
					},
				},
			},
		}, {
			input: []arg[N]{
				{ctx, 1, alice},
				{ctx, 1, bob},
				// These will exceed cardinality limit.
				{ctx, 1, carol},
				{ctx, 1, dave},
			},
			remove: []arg[N]{},
			expect: output{
				n: 3,
				agg: metricdata.Gauge[N]{
					DataPoints: []metricdata.DataPoint[N]{
						{
							Attributes: fltrAlice,
							StartTime:  y2kPlus(11),
							Time:       y2kPlus(12),
							Value:      1,
						},
						{
							Attributes: fltrBob,
							StartTime:  y2kPlus(11),
							Time:       y2kPlus(13),
							Value:      1,
						},
						{
							Attributes: overflowSet,
							StartTime:  y2kPlus(11),
							Time:       y2kPlus(15),
							Value:      1,
						},
					},
				},
			},
		},
	})
}

func testCumulativePrecomputedLastValue[N int64 | float64]() func(*testing.T) {
	in, remove, out := Builder[N]{
		Temporality:      metricdata.CumulativeTemporality,
		Filter:           attrFltr,
		AggregationLimit: 3,
	}.PrecomputedLastValue()
	ctx := context.Background()
	return test[N](in, remove, out, []teststep[N]{
		{
			// Empty output if nothing is measured.
			input:  []arg[N]{},
			expect: output{n: 0, agg: metricdata.Gauge[N]{}},
		}, {
			input: []arg[N]{
				{ctx, 1, alice},
				{ctx, -1, bob},
				{ctx, 1, fltrAlice},
				{ctx, 2, alice},
				{ctx, -10, bob},
			},
			remove: []arg[N]{},
			expect: output{
				n: 2,
				agg: metricdata.Gauge[N]{
					DataPoints: []metricdata.DataPoint[N]{
						{
							Attributes: fltrAlice,
							StartTime:  y2kPlus(0),
							Time:       y2kPlus(4),
							Value:      2,
						},
						{
							Attributes: fltrBob,
							StartTime:  y2kPlus(0),
							Time:       y2kPlus(5),
							Value:      -10,
						},
					},
				},
			},
		}, {
			// Everything resets, do not report old measurements.
			input:  []arg[N]{},
			remove: []arg[N]{},
			expect: output{n: 0, agg: metricdata.Gauge[N]{}},
		}, {
			input: []arg[N]{
				{ctx, 10, alice},
				{ctx, 3, bob},
			},
			remove: []arg[N]{},
			expect: output{
				n: 2,
				agg: metricdata.Gauge[N]{
					DataPoints: []metricdata.DataPoint[N]{
						{
							Attributes: fltrAlice,
							StartTime:  y2kPlus(0),
							Time:       y2kPlus(6),
							Value:      10,
						},
						{
							Attributes: fltrBob,
							StartTime:  y2kPlus(0),
							Time:       y2kPlus(7),
							Value:      3,
						},
					},
				},
			},
		}, {
			input: []arg[N]{
				{ctx, 1, alice},
				{ctx, 1, bob},
				// These will exceed cardinality limit.
				{ctx, 1, carol},
				{ctx, 1, dave},
			},
			remove: []arg[N]{},
			expect: output{
				n: 3,
				agg: metricdata.Gauge[N]{
					DataPoints: []metricdata.DataPoint[N]{
						{
							Attributes: fltrAlice,
							StartTime:  y2kPlus(0),
							Time:       y2kPlus(8),
							Value:      1,
						},
						{
							Attributes: fltrBob,
							StartTime:  y2kPlus(0),
							Time:       y2kPlus(9),
							Value:      1,
						},
						{
							Attributes: overflowSet,
							StartTime:  y2kPlus(0),
							Time:       y2kPlus(11),
							Value:      1,
						},
					},
				},
			},
		},
	})
}

func BenchmarkLastValue(b *testing.B) {
	b.Run("Int64", benchmarkAggregate(Builder[int64]{}.PrecomputedLastValue))
	b.Run("Float64", benchmarkAggregate(Builder[float64]{}.PrecomputedLastValue))
}
