package dto_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/BurntSushi/toml"

	"github.com/thomaspoignant/go-feature-flag/internal/flag"

	"github.com/thomaspoignant/go-feature-flag/testutils/testconvert"

	"github.com/thomaspoignant/go-feature-flag/internal/dto"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/stretchr/testify/assert"
)

func TestRollout_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		fileType string
		want     dto.Rollout
		wantErr  assert.ErrorAssertionFunc
	}{
		{
			name:     "Should unmarshal valid JSON file",
			input:    "../../testdata/internal/dto/rollout.json",
			fileType: "json",
			want: dto.Rollout{
				V1Rollout: dto.V1Rollout{
					Scheduled: &[]flag.ScheduledStep{
						{
							Date: testconvert.Time(time.Date(2021, time.February, 1, 10, 10, 10, 10, time.UTC)),
							InternalFlag: flag.InternalFlag{
								Variations: &map[string]*interface{}{
									"VariationDefault": testconvert.Interface(true),
								},
							},
						},
						{
							Date: testconvert.Time(time.Date(2021, time.February, 2, 10, 10, 10, 10, time.UTC)),
							InternalFlag: flag.InternalFlag{
								Variations: &map[string]*interface{}{
									"VariationDefault": testconvert.Interface(false),
								},
							},
						},
					},
				},
				V0Rollout: dto.V0Rollout{Scheduled: &dto.ScheduledRolloutV0{}},
				CommonRollout: dto.CommonRollout{
					Experimentation: &dto.ExperimentationDto{
						Start: testconvert.Time(time.Date(2021, time.February, 1, 10, 10, 10, 10, time.UTC)),
						End:   testconvert.Time(time.Date(2021, time.March, 1, 10, 10, 10, 10, time.UTC)),
					},
					Progressive: nil,
				},
			},
			wantErr: assert.NoError,
		},
		{
			name:     "Should unmarshal valid TOML file",
			input:    "../../testdata/internal/dto/rollout.toml",
			fileType: "toml",
			want: dto.Rollout{
				V1Rollout: dto.V1Rollout{
					Scheduled: &[]flag.ScheduledStep{
						{
							Date: testconvert.Time(time.Date(2021, time.February, 1, 10, 10, 10, 10, time.UTC)),
							InternalFlag: flag.InternalFlag{
								Variations: &map[string]*interface{}{
									"VariationDefault": testconvert.Interface(true),
								},
							},
						},
						{
							Date: testconvert.Time(time.Date(2021, time.February, 2, 10, 10, 10, 10, time.UTC)),
							InternalFlag: flag.InternalFlag{
								Variations: &map[string]*interface{}{
									"VariationDefault": testconvert.Interface(false),
								},
							},
						},
					},
				},
				V0Rollout: dto.V0Rollout{Scheduled: &dto.ScheduledRolloutV0{}},
				CommonRollout: dto.CommonRollout{
					Experimentation: &dto.ExperimentationDto{
						Start: testconvert.Time(time.Date(2021, time.February, 1, 10, 10, 10, 10, time.UTC)),
						End:   testconvert.Time(time.Date(2021, time.March, 1, 10, 10, 10, 10, time.UTC)),
					},
					Progressive: nil,
				},
			},
			wantErr: assert.NoError,
		},
		{
			name:     "Should unmarshal valid YAML file",
			input:    "../../testdata/internal/dto/rollout.yaml",
			fileType: "yaml",
			want: dto.Rollout{
				V1Rollout: dto.V1Rollout{
					Scheduled: &[]flag.ScheduledStep{
						{
							Date: testconvert.Time(time.Date(2021, time.February, 1, 10, 10, 10, 10, time.UTC)),
							InternalFlag: flag.InternalFlag{
								Variations: &map[string]*interface{}{
									"VariationDefault": testconvert.Interface(true),
								},
							},
						},
						{
							Date: testconvert.Time(time.Date(2021, time.February, 2, 10, 10, 10, 10, time.UTC)),
							InternalFlag: flag.InternalFlag{
								Variations: &map[string]*interface{}{
									"VariationDefault": testconvert.Interface(false),
								},
							},
						},
					},
				},
				V0Rollout: dto.V0Rollout{Scheduled: &dto.ScheduledRolloutV0{}},
				CommonRollout: dto.CommonRollout{
					Experimentation: &dto.ExperimentationDto{
						Start: testconvert.Time(time.Date(2021, time.February, 1, 10, 10, 10, 10, time.UTC)),
						End:   testconvert.Time(time.Date(2021, time.March, 1, 10, 10, 10, 10, time.UTC)),
					},
					Progressive: nil,
				},
			},
			wantErr: assert.NoError,
		},
		{
			name:     "Should unmarshal valid JSON file with v0 rollout",
			input:    "../../testdata/internal/dto/rollout_v0.json",
			fileType: "JSON",
			want: dto.Rollout{
				V0Rollout: dto.V0Rollout{Scheduled: &dto.ScheduledRolloutV0{
					Steps: []dto.ScheduledStepV0{
						{
							DTO: dto.DTO{
								DTOv0: dto.DTOv0{
									Rule:       testconvert.String("beta eq \"true\""),
									Percentage: testconvert.Float64(100),
								},
							},
							Date: testconvert.Time(time.Date(2021, time.February, 2, 10, 10, 10, 10, time.UTC)),
						},
						{
							DTO: dto.DTO{
								DTOv0: dto.DTOv0{
									Rule: testconvert.String("beta eq \"false\""),
								},
							},
							Date: testconvert.Time(time.Date(2021, time.February, 2, 10, 10, 10, 10, time.UTC)),
						},
					},
				}},
				CommonRollout: dto.CommonRollout{
					Experimentation: &dto.ExperimentationDto{
						Start: testconvert.Time(time.Date(2021, time.February, 1, 10, 10, 10, 10, time.UTC)),
						End:   testconvert.Time(time.Date(2021, time.March, 1, 10, 10, 10, 10, time.UTC)),
					},
					Progressive: nil,
				},
			},
			wantErr: assert.NoError,
		},
		{
			name:     "Should unmarshal valid TOML file with v0 rollout",
			input:    "../../testdata/internal/dto/rollout_v0.toml",
			fileType: "TOML",
			want: dto.Rollout{
				V0Rollout: dto.V0Rollout{Scheduled: &dto.ScheduledRolloutV0{
					Steps: []dto.ScheduledStepV0{
						{
							DTO: dto.DTO{
								DTOv0: dto.DTOv0{
									Rule:       testconvert.String("beta eq \"true\""),
									Percentage: testconvert.Float64(100),
								},
							},
							Date: testconvert.Time(time.Date(2021, time.February, 2, 10, 10, 10, 10, time.UTC)),
						},
						{
							DTO: dto.DTO{
								DTOv0: dto.DTOv0{
									Rule: testconvert.String("beta eq \"false\""),
								},
							},
							Date: testconvert.Time(time.Date(2021, time.February, 2, 10, 10, 10, 10, time.UTC)),
						},
					},
				}},
				CommonRollout: dto.CommonRollout{
					Experimentation: &dto.ExperimentationDto{
						Start: testconvert.Time(time.Date(2021, time.February, 1, 10, 10, 10, 10, time.UTC)),
						End:   testconvert.Time(time.Date(2021, time.March, 1, 10, 10, 10, 10, time.UTC)),
					},
					Progressive: nil,
				},
			},
			wantErr: assert.NoError,
		},
		{
			name:     "Should unmarshal valid YAML file with v0 rollout",
			input:    "../../testdata/internal/dto/rollout_v0.yaml",
			fileType: "YAML",
			want: dto.Rollout{
				V0Rollout: dto.V0Rollout{Scheduled: &dto.ScheduledRolloutV0{
					Steps: []dto.ScheduledStepV0{
						{
							DTO: dto.DTO{
								DTOv0: dto.DTOv0{
									Rule:       testconvert.String("beta eq \"true\""),
									Percentage: testconvert.Float64(100),
								},
							},
							Date: testconvert.Time(time.Date(2021, time.February, 2, 10, 10, 10, 10, time.UTC)),
						},
						{
							DTO: dto.DTO{
								DTOv0: dto.DTOv0{
									Rule: testconvert.String("beta eq \"false\""),
								},
							},
							Date: testconvert.Time(time.Date(2021, time.February, 2, 10, 10, 10, 10, time.UTC)),
						},
					},
				}},
				CommonRollout: dto.CommonRollout{
					Experimentation: &dto.ExperimentationDto{
						Start: testconvert.Time(time.Date(2021, time.February, 1, 10, 10, 10, 10, time.UTC)),
						End:   testconvert.Time(time.Date(2021, time.March, 1, 10, 10, 10, 10, time.UTC)),
					},
					Progressive: nil,
				},
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			var r dto.Rollout
			content, err := os.ReadFile(tt.input)
			assert.NoError(t, err, "impossible to find input test file %v", tt.input)

			switch strings.ToLower(tt.fileType) {
			case "toml":
				_, err = toml.Decode(string(content), &r)
			case "json":
				err = json.Unmarshal(content, &r)
			case "yaml":
				err = yaml.Unmarshal(content, &r)
			default:
				panic("not expected")
			}
			tt.wantErr(t, err)
			assert.Equal(t, tt.want, r, cmp.Diff(tt.want, r))
		})
	}
}
