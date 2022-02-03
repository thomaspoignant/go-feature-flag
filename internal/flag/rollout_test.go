package flag_test

//func TestExperimentation_String(t *testing.T) {
//	type fields struct {
//		StartDate *time.Time
//		EndDate   *time.Time
//		Start     *time.Time
//		End       *time.Time
//	}
//	tests := []struct {
//		name   string
//		fields fields
//		want   string
//	}{
//		{
//			name: "both dates",
//			fields: fields{
//				Start: testconvert.Time(time.Unix(1095379400, 0)),
//				End:   testconvert.Time(time.Unix(1095379500, 0)),
//			},
//			want: "start:[2004-09-17T00:03:20Z] end:[2004-09-17T00:05:00Z]",
//		},
//		{
//			name: "only start date",
//			fields: fields{
//				Start: testconvert.Time(time.Unix(1095379400, 0)),
//			},
//			want: "start:[2004-09-17T00:03:20Z]",
//		},
//		{
//			name: "only end date",
//			fields: fields{
//				End: testconvert.Time(time.Unix(1095379500, 0)),
//			},
//			want: "end:[2004-09-17T00:05:00Z]",
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			e := flag.Experimentation{
//				End:   tt.fields.End,
//				Start: tt.fields.Start,
//			}
//			got := e.String()
//			assert.Equal(t, tt.want, got)
//		})
//	}
//}
