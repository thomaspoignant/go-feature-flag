package dto_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/ffuser"
	"github.com/thomaspoignant/go-feature-flag/internal/dto"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	"github.com/thomaspoignant/go-feature-flag/testutils/flagv1"
	"github.com/thomaspoignant/go-feature-flag/testutils/testconvert"
	"testing"
	"time"
)

func Test_toto(t *testing.T) {
	u := ffuser.NewUser("yo")
	flagName := "yo"
	d := dto.DTO{
		DTOv0: dto.DTOv0{
			True:       testconvert.Interface("true"),
			False:      testconvert.Interface("false"),
			Default:    testconvert.Interface("default"),
			Percentage: testconvert.Float64(100),
			Rollout: &dto.RolloutV0{
				Scheduled: &dto.ScheduledRolloutV0{Steps: []dto.ScheduledStepV0{
					{
						DTOv0: dto.DTOv0{
							Percentage: testconvert.Float64(10),
						},
						Date: testconvert.Time(time.Now().Add(-2 * time.Second)),
					},
					{
						DTOv0: dto.DTOv0{
							Rule: testconvert.String("key eq \"yo\""),
						},
						Date: testconvert.Time(time.Now().Add(-2 * time.Second)),
					},
					{
						DTOv0: dto.DTOv0{
							True: testconvert.Interface("newValue"),
						},
						Date: testconvert.Time(time.Now().Add(-2 * time.Second)),
					},
					{
						DTOv0: dto.DTOv0{
							Percentage: testconvert.Float64(10),
						},
						Date: testconvert.Time(time.Now().Add(-2 * time.Second)),
					},
					{
						DTOv0: dto.DTOv0{
							Rule:       testconvert.String(""),
							Percentage: testconvert.Float64(100),
						},
						Date: testconvert.Time(time.Now().Add(-2 * time.Second)),
					},
					{
						DTOv0: dto.DTOv0{
							Disable: testconvert.Bool(true),
						},
						Date: testconvert.Time(time.Now().Add(-2 * time.Second)),
					},
					{
						DTOv0: dto.DTOv0{
							Disable:    testconvert.Bool(false),
							Rule:       testconvert.String("anonymous eq false"),
							Percentage: testconvert.Float64(100),
						},
						Date: testconvert.Time(time.Now().Add(-2 * time.Second)),
					},
					{
						DTOv0: dto.DTOv0{
							Rollout: &dto.RolloutV0{
								Experimentation: &dto.ExperimentationV0{
									Start: testconvert.Time(time.Now().Add(-2 * time.Second)),
									End:   testconvert.Time(time.Now().Add(2 * time.Second)),
								},
							},
						},
						Date: testconvert.Time(time.Now().Add(-2 * time.Second)),
					},
					{
						DTOv0: dto.DTOv0{
							Rollout: &dto.RolloutV0{
								Experimentation: &dto.ExperimentationV0{
									End: testconvert.Time(time.Now().Add(-2 * time.Second)),
								},
							},
						},
						Date: testconvert.Time(time.Now().Add(-2 * time.Second)),
					},
				}},
			},
		},
	}

	c := d.Convert()
	n, _ := c.Value(flagName, u, flag.EvaluationContext{})

	e := flagv1.ConvertDtoToV1(d.DTOv0)
	m, _ := e.Value(flagName, u, flag.EvaluationContext{})

	fmt.Println(m)
	assert.Equal(t, m, n)
	//assert.Equal(t, n1, m1)

}
