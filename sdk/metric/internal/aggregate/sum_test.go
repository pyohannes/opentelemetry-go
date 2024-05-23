// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package aggregate // import "go.opentelemetry.io/otel/sdk/metric/internal/aggregate"

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestSum(t *testing.T) {
	c := new(clock)
	t.Cleanup(c.Register())

	t.Run("Int64/DeltaSum", testDeltaSum[int64]())
	c.Reset()

	t.Run("Float64/DeltaSum", testDeltaSum[float64]())
	c.Reset()

	t.Run("Int64/CumulativeSum", testCumulativeSum[int64]())
	c.Reset()

	t.Run("Float64/CumulativeSum", testCumulativeSum[float64]())
	c.Reset()

	t.Run("Int64/DeltaPrecomputedSum", testDeltaPrecomputedSum[int64]())
	c.Reset()

	t.Run("Float64/DeltaPrecomputedSum", testDeltaPrecomputedSum[float64]())
	c.Reset()

	t.Run("Int64/CumulativePrecomputedSum", testCumulativePrecomputedSum[int64]())
	c.Reset()

	t.Run("Float64/CumulativePrecomputedSum", testCumulativePrecomputedSum[float64]())
}

func testDeltaSum[N int64 | float64]() func(t *testing.T) {
	mono := false
	in, remove, out := Builder[N]{
		Temporality:      metricdata.DeltaTemporality,
		Filter:           attrFltr,
		AggregationLimit: 3,
	}.Sum(mono)
	ctx := context.Background()
	return test[N](in, remove, out, []teststep[N]{
		{
			input: []arg[N]{},
			expect: output{
				n: 0,
				agg: metricdata.Sum[N]{
					IsMonotonic: mono,
					Temporality: metricdata.DeltaTemporality,
					DataPoints:  []metricdata.DataPoint[N]{},
				},
			},
		},
		{
			input: []arg[N]{
				{ctx, 1, alice},
				{ctx, -1, bob},
				{ctx, 1, alice},
				{ctx, 2, alice},
				{ctx, -10, bob},
			},
			remove: []arg[N]{},
			expect: output{
				n: 2,
				agg: metricdata.Sum[N]{
					IsMonotonic: mono,
					Temporality: metricdata.DeltaTemporality,
					DataPoints: []metricdata.DataPoint[N]{
						{
							Attributes: fltrAlice,
							StartTime:  y2kPlus(1),
							Time:       y2kPlus(7),
							Value:      4,
						},
						{
							Attributes: fltrBob,
							StartTime:  y2kPlus(1),
							Time:       y2kPlus(7),
							Value:      -11,
						},
					},
				},
			},
		},
		{
			input: []arg[N]{
				{ctx, 10, alice},
				{ctx, 3, bob},
			},
			remove: []arg[N]{},
			expect: output{
				n: 2,
				agg: metricdata.Sum[N]{
					IsMonotonic: mono,
					Temporality: metricdata.DeltaTemporality,
					DataPoints: []metricdata.DataPoint[N]{
						{
							Attributes: fltrAlice,
							StartTime:  y2kPlus(7),
							Time:       y2kPlus(10),
							Value:      10,
						},
						{
							Attributes: fltrBob,
							StartTime:  y2kPlus(7),
							Time:       y2kPlus(10),
							Value:      3,
						},
					},
				},
			},
		},
		{
			input: []arg[N]{},
			// Delta sums are expected to reset.
			remove: []arg[N]{},
			expect: output{
				n: 0,
				agg: metricdata.Sum[N]{
					IsMonotonic: mono,
					Temporality: metricdata.DeltaTemporality,
					DataPoints:  []metricdata.DataPoint[N]{},
				},
			},
		},
		{
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
				agg: metricdata.Sum[N]{
					IsMonotonic: mono,
					Temporality: metricdata.DeltaTemporality,
					DataPoints: []metricdata.DataPoint[N]{
						{
							Attributes: fltrAlice,
							StartTime:  y2kPlus(11),
							Time:       y2kPlus(16),
							Value:      1,
						},
						{
							Attributes: fltrBob,
							StartTime:  y2kPlus(11),
							Time:       y2kPlus(16),
							Value:      1,
						},
						{
							Attributes: overflowSet,
							StartTime:  y2kPlus(11),
							Time:       y2kPlus(16),
							Value:      2,
						},
					},
				},
			},
		},
	})
}

func testCumulativeSum[N int64 | float64]() func(t *testing.T) {
	mono := false
	in, remove, out := Builder[N]{
		Temporality:      metricdata.CumulativeTemporality,
		Filter:           attrFltr,
		AggregationLimit: 3,
	}.Sum(mono)
	ctx := context.Background()
	return test[N](in, remove, out, []teststep[N]{
		{
			input: []arg[N]{},
			remove: []arg[N]{},
			expect: output{
				n: 0,
				agg: metricdata.Sum[N]{
					IsMonotonic: mono,
					Temporality: metricdata.CumulativeTemporality,
					DataPoints:  []metricdata.DataPoint[N]{},
				},
			},
		},
		{
			input: []arg[N]{
				{ctx, 1, alice},
				{ctx, -1, bob},
				{ctx, 1, alice},
				{ctx, 2, alice},
				{ctx, -10, bob},
			},
			remove: []arg[N]{},
			expect: output{
				n: 2,
				agg: metricdata.Sum[N]{
					IsMonotonic: mono,
					Temporality: metricdata.CumulativeTemporality,
					DataPoints: []metricdata.DataPoint[N]{
						{
							Attributes: fltrAlice,
							StartTime:  y2kPlus(0),
							Time:       y2kPlus(7),
							Value:      4,
						},
						{
							Attributes: fltrBob,
							StartTime:  y2kPlus(0),
							Time:       y2kPlus(7),
							Value:      -11,
						},
					},
				},
			},
		},
		{
			input: []arg[N]{
				{ctx, 10, alice},
				{ctx, 3, bob},
			},
			remove: []arg[N]{},
			expect: output{
				n: 2,
				agg: metricdata.Sum[N]{
					IsMonotonic: mono,
					Temporality: metricdata.CumulativeTemporality,
					DataPoints: []metricdata.DataPoint[N]{
						{
							Attributes: fltrAlice,
							StartTime:  y2kPlus(0),
							Time:       y2kPlus(10),
							Value:      14,
						},
						{
							Attributes: fltrBob,
							StartTime:  y2kPlus(0),
							Time:       y2kPlus(10),
							Value:      -8,
						},
					},
				},
			},
		},
		{
			input: []arg[N]{
				// These will exceed cardinality limit.
				{ctx, 1, carol},
				{ctx, 1, dave},
			},
			remove: []arg[N]{},
			expect: output{
				n: 3,
				agg: metricdata.Sum[N]{
					IsMonotonic: mono,
					Temporality: metricdata.CumulativeTemporality,
					DataPoints: []metricdata.DataPoint[N]{
						{
							Attributes: fltrAlice,
							StartTime:  y2kPlus(0),
							Time:       y2kPlus(13),
							Value:      14,
						},
						{
							Attributes: fltrBob,
							StartTime:  y2kPlus(0),
							Time:       y2kPlus(13),
							Value:      -8,
						},
						{
							Attributes: overflowSet,
							StartTime:  y2kPlus(0),
							Time:       y2kPlus(13),
							Value:      2,
						},
					},
				},
			},
		},
		{
			input: []arg[N]{},
			remove: []arg[N]{
				{ctx, 0, fltrBob},
			},
			expect: output{
				n: 3,
				agg: metricdata.Sum[N]{
					IsMonotonic: mono,
					Temporality: metricdata.CumulativeTemporality,
					DataPoints: []metricdata.DataPoint[N]{
						{
							Attributes: fltrAlice,
							StartTime:  y2kPlus(0),
							Time:       y2kPlus(14),
							Value:      14,
						},
						{
							Attributes: fltrBob,
							StartTime:  y2kPlus(0),
							Time:       y2kPlus(14),
							NoRecordedValue: true,
						},
						{
							Attributes: overflowSet,
							StartTime:  y2kPlus(0),
							Time:       y2kPlus(14),
							Value:      2,
						},
					},
				},
			},
		},
		{
			input: []arg[N]{},
			remove: []arg[N]{},
			expect: output{
				n: 2,
				agg: metricdata.Sum[N]{
					IsMonotonic: mono,
					Temporality: metricdata.CumulativeTemporality,
					DataPoints: []metricdata.DataPoint[N]{
						{
							Attributes: fltrAlice,
							StartTime:  y2kPlus(0),
							Time:       y2kPlus(15),
							Value:      14,
						},
						{
							Attributes: overflowSet,
							StartTime:  y2kPlus(0),
							Time:       y2kPlus(15),
							Value:      2,
						},
					},
				},
			},
		},
	})
}

func testDeltaPrecomputedSum[N int64 | float64]() func(t *testing.T) {
	mono := false
	in, remove, out := Builder[N]{
		Temporality:      metricdata.DeltaTemporality,
		Filter:           attrFltr,
		AggregationLimit: 3,
	}.PrecomputedSum(mono)
	ctx := context.Background()
	return test[N](in, remove, out, []teststep[N]{
		{
			input: []arg[N]{},
			remove: []arg[N]{},
			expect: output{
				n: 0,
				agg: metricdata.Sum[N]{
					IsMonotonic: mono,
					Temporality: metricdata.DeltaTemporality,
					DataPoints:  []metricdata.DataPoint[N]{},
				},
			},
		},
		{
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
				agg: metricdata.Sum[N]{
					IsMonotonic: mono,
					Temporality: metricdata.DeltaTemporality,
					DataPoints: []metricdata.DataPoint[N]{
						{
							Attributes: fltrAlice,
							StartTime:  y2kPlus(1),
							Time:       y2kPlus(7),
							Value:      4,
						},
						{
							Attributes: fltrBob,
							StartTime:  y2kPlus(1),
							Time:       y2kPlus(7),
							Value:      -11,
						},
					},
				},
			},
		},
		{
			input: []arg[N]{
				{ctx, 1, fltrAlice},
				{ctx, 10, alice},
				{ctx, 3, bob},
			},
			remove: []arg[N]{},
			expect: output{
				n: 2,
				agg: metricdata.Sum[N]{
					IsMonotonic: mono,
					Temporality: metricdata.DeltaTemporality,
					DataPoints: []metricdata.DataPoint[N]{
						{
							Attributes: fltrAlice,
							StartTime:  y2kPlus(7),
							Time:       y2kPlus(11),
							Value:      7,
						},
						{
							Attributes: fltrBob,
							StartTime:  y2kPlus(7),
							Time:       y2kPlus(11),
							Value:      14,
						},
					},
				},
			},
		},
		{
			input: []arg[N]{},
			// Precomputed sums are expected to reset.
			remove: []arg[N]{},
			expect: output{
				n: 0,
				agg: metricdata.Sum[N]{
					IsMonotonic: mono,
					Temporality: metricdata.DeltaTemporality,
					DataPoints:  []metricdata.DataPoint[N]{},
				},
			},
		},
		{
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
				agg: metricdata.Sum[N]{
					IsMonotonic: mono,
					Temporality: metricdata.DeltaTemporality,
					DataPoints: []metricdata.DataPoint[N]{
						{
							Attributes: fltrAlice,
							StartTime:  y2kPlus(12),
							Time:       y2kPlus(17),
							Value:      1,
						},
						{
							Attributes: fltrBob,
							StartTime:  y2kPlus(12),
							Time:       y2kPlus(17),
							Value:      1,
						},
						{
							Attributes: overflowSet,
							StartTime:  y2kPlus(12),
							Time:       y2kPlus(17),
							Value:      2,
						},
					},
				},
			},
		},
	})
}

func testCumulativePrecomputedSum[N int64 | float64]() func(t *testing.T) {
	mono := false
	in, remove, out := Builder[N]{
		Temporality:      metricdata.CumulativeTemporality,
		Filter:           attrFltr,
		AggregationLimit: 3,
	}.PrecomputedSum(mono)
	ctx := context.Background()
	return test[N](in, remove, out, []teststep[N]{
		{
			input: []arg[N]{},
			remove: []arg[N]{},
			expect: output{
				n: 0,
				agg: metricdata.Sum[N]{
					IsMonotonic: mono,
					Temporality: metricdata.CumulativeTemporality,
					DataPoints:  []metricdata.DataPoint[N]{},
				},
			},
		},
		{
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
				agg: metricdata.Sum[N]{
					IsMonotonic: mono,
					Temporality: metricdata.CumulativeTemporality,
					DataPoints: []metricdata.DataPoint[N]{
						{
							Attributes: fltrAlice,
							StartTime:  y2kPlus(0),
							Time:       y2kPlus(7),
							Value:      4,
						},
						{
							Attributes: fltrBob,
							StartTime:  y2kPlus(0),
							Time:       y2kPlus(7),
							Value:      -11,
						},
					},
				},
			},
		},
		{
			input: []arg[N]{
				{ctx, 1, fltrAlice},
				{ctx, 10, alice},
				{ctx, 3, bob},
			},
			remove: []arg[N]{},
			expect: output{
				n: 2,
				agg: metricdata.Sum[N]{
					IsMonotonic: mono,
					Temporality: metricdata.CumulativeTemporality,
					DataPoints: []metricdata.DataPoint[N]{
						{
							Attributes: fltrAlice,
							StartTime:  y2kPlus(0),
							Time:       y2kPlus(11),
							Value:      11,
						},
						{
							Attributes: fltrBob,
							StartTime:  y2kPlus(0),
							Time:       y2kPlus(11),
							Value:      3,
						},
					},
				},
			},
		},
		{
			input: []arg[N]{},
			// Precomputed sums are expected to reset.
			remove: []arg[N]{},
			expect: output{
				n: 0,
				agg: metricdata.Sum[N]{
					IsMonotonic: mono,
					Temporality: metricdata.CumulativeTemporality,
					DataPoints:  []metricdata.DataPoint[N]{},
				},
			},
		},
		{
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
				agg: metricdata.Sum[N]{
					IsMonotonic: mono,
					Temporality: metricdata.CumulativeTemporality,
					DataPoints: []metricdata.DataPoint[N]{
						{
							Attributes: fltrAlice,
							StartTime:  y2kPlus(0),
							Time:       y2kPlus(17),
							Value:      1,
						},
						{
							Attributes: fltrBob,
							StartTime:  y2kPlus(0),
							Time:       y2kPlus(17),
							Value:      1,
						},
						{
							Attributes: overflowSet,
							StartTime:  y2kPlus(0),
							Time:       y2kPlus(17),
							Value:      2,
						},
					},
				},
			},
		},
	})
}

func BenchmarkSum(b *testing.B) {
	// The monotonic argument is only used to annotate the Sum returned from
	// the Aggregation method. It should not have an effect on operational
	// performance, therefore, only monotonic=false is benchmarked here.
	b.Run("Int64/Cumulative", benchmarkAggregate(func() (Measure[int64], Remove, ComputeAggregation) {
		return Builder[int64]{
			Temporality: metricdata.CumulativeTemporality,
		}.Sum(false)
	}))
	b.Run("Int64/Delta", benchmarkAggregate(func() (Measure[int64], Remove, ComputeAggregation) {
		return Builder[int64]{
			Temporality: metricdata.DeltaTemporality,
		}.Sum(false)
	}))
	b.Run("Float64/Cumulative", benchmarkAggregate(func() (Measure[float64], Remove, ComputeAggregation) {
		return Builder[float64]{
			Temporality: metricdata.CumulativeTemporality,
		}.Sum(false)
	}))
	b.Run("Float64/Delta", benchmarkAggregate(func() (Measure[float64], Remove, ComputeAggregation) {
		return Builder[float64]{
			Temporality: metricdata.DeltaTemporality,
		}.Sum(false)
	}))

	b.Run("Precomputed/Int64/Cumulative", benchmarkAggregate(func() (Measure[int64], Remove, ComputeAggregation) {
		return Builder[int64]{
			Temporality: metricdata.CumulativeTemporality,
		}.PrecomputedSum(false)
	}))
	b.Run("Precomputed/Int64/Delta", benchmarkAggregate(func() (Measure[int64], Remove, ComputeAggregation) {
		return Builder[int64]{
			Temporality: metricdata.DeltaTemporality,
		}.PrecomputedSum(false)
	}))
	b.Run("Precomputed/Float64/Cumulative", benchmarkAggregate(func() (Measure[float64], Remove, ComputeAggregation) {
		return Builder[float64]{
			Temporality: metricdata.CumulativeTemporality,
		}.PrecomputedSum(false)
	}))
	b.Run("Precomputed/Float64/Delta", benchmarkAggregate(func() (Measure[float64], Remove, ComputeAggregation) {
		return Builder[float64]{
			Temporality: metricdata.DeltaTemporality,
		}.PrecomputedSum(false)
	}))
}
