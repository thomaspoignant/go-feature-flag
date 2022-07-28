package flag_test

// func TestScheduledRollout_String(t *testing.T) {
//	type fields struct {
//		Steps []flag.ScheduledStep
//	}
//	tests := []struct {
//		name   string
//		fields fields
//		want   string
//	}{
//		{
//			name: "Simple 2 steps",
//			want: "[2021-02-01T10:10:10Z: Version:[1.10]],[2021-02-02T10:10:10Z: Variations:[A=yo,B=y1]]",
//			fields: fields{Steps: []flag.ScheduledStep{
//				{
//					InternalFlag: flag.InternalFlag{
//						Version: testconvert.String(fmt.Sprintf("%.2f", 1.1)),
//					},
//					Date: testconvert.Time(time.Date(2021, time.February, 1, 10, 10, 10, 10, time.UTC)),
//				},
//				{
//					InternalFlag: flag.InternalFlag{
//						Variations: &map[string]*interface{}{
//							"A": testconvert.Interface("yo"),
//							"B": testconvert.Interface("y1"),
//						},
//					},
//					Date: testconvert.Time(time.Date(2021, time.February, 2, 10, 10, 10, 10, time.UTC)),
//				},
//			}},
//		},
//		{
//			name: "Complex steps",
//			want: "[2021-02-01T10:10:10Z: Version:[1.10]],[2021-02-02T10:10:10Z: Variations:[A=yo,B=y1], " +
//				"DefaultRule:[query:[key eq \"toto\"], percentages:" +
//				"[A=10.00,B=90.00], progressiveRollout:[Initial:[Variation:[A], Percentage:[10], Date:[2021-02-01T10:10:10Z]]," +
//				" End:[Variation:[B], Percentage:[90], Date:[2021-02-04T10:10:10Z]]]]]",
//			fields: fields{Steps: []flag.ScheduledStep{
//				{
//					InternalFlag: flag.InternalFlag{
//						Version: testconvert.String(fmt.Sprintf("%.2f", 1.1)),
//					},
//					Date: testconvert.Time(time.Date(2021, time.February, 1, 10, 10, 10, 10, time.UTC)),
//				},
//				{
//					InternalFlag: flag.InternalFlag{
//						Variations: &map[string]*interface{}{
//							"A": testconvert.Interface("yo"),
//							"B": testconvert.Interface("y1"),
//						},
//						DefaultRule: &flag.Rule{
//							Query: testconvert.String("key eq \"toto\""),
//							Percentages: &map[string]float64{
//								"A": 10,
//								"B": 90,
//							},
//							ProgressiveRollout: &flag.ProgressiveRollout{
//								Initial: &flag.ProgressiveRolloutStep{
//									Variation:  testconvert.String("A"),
//									Percentage: 10,
//									Date:       testconvert.Time(time.Date(2021, time.February, 1, 10, 10, 10, 10, time.UTC)),
//								},
//								End: &flag.ProgressiveRolloutStep{
//									Variation:  testconvert.String("B"),
//									Percentage: 90,
//									Date:       testconvert.Time(time.Date(2021, time.February, 4, 10, 10, 10, 10, time.UTC)),
//								},
//							},
//						},
//					},
//					Date: testconvert.Time(time.Date(2021, time.February, 2, 10, 10, 10, 10, time.UTC)),
//				},
//			}},
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			s := flag.ScheduledRollout{
//				Steps: tt.fields.Steps,
//			}
//			assert.Equalf(t, tt.want, s.String(), "String()")
//		})
//	}
//}
