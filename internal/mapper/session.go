package mapper

import (
	"time"

	pb "github.com/barsboldb/ascend-backend/gen/session"
	"github.com/barsboldb/ascend-backend/internal/model"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type exerciseWithSets struct {
  Exercise model.Exercise
  Sets     []model.ExerciseSet
}

func SessionToPB(session *model.Session) (*pb.SessionWithExercises) {
  grouped := make(map[uuid.UUID]*exerciseWithSets)
  var order []uuid.UUID

  for _, set := range session.ExerciseSets {
    if _, ok := grouped[set.ExerciseID]; !ok {
      grouped[set.ExerciseID] = &exerciseWithSets{Exercise: set.Exercise}
      order = append(order, set.ExerciseID)
    }
    grouped[set.ExerciseID].Sets = append(grouped[set.ExerciseID].Sets, set)
  }

  pbExercises := make([]*pb.SessionExercise, len(order))
  for i, exID := range order {
    g := grouped[exID]
    pbSets := make([]*pb.ExerciseSet, len(g.Sets))
    for j, set := range g.Sets {
      pbSets[j] = &pb.ExerciseSet{
        SetNumber: set.SetNumber,
        WeightKg:  float32(set.WeightKg),
        Reps:      set.Reps,
        Failure:   set.Failure,
      }
    }
    pbExercises[i] = &pb.SessionExercise{
      ExerciseId:   exID.String(),
      ExerciseName: g.Exercise.Name,
      Sets:         pbSets,
    }
  }

	resp := &pb.SessionWithExercises{
		Id:           session.ID.String(),
		ProgramDayId: session.ProgramDayID.String(),
		WeekNumber:   session.WeekNumber,
		StartedAt:    timestamppb.New(session.StartedAt),
		Exercises:    pbExercises,
	}
  return resp
}

func PBToSession(sessionPB *pb.CreateSessionRequest) (*model.Session) {
  programDayId := uuid.MustParse(sessionPB.ProgramDayId)
  var endedAt *time.Time 
  if sessionPB.EndedAt != nil {
    t := sessionPB.EndedAt.AsTime();
    endedAt = &t
  }

  sets := make([]model.ExerciseSet, len(sessionPB.Exercises))
  for i, e := range sessionPB.Exercises {
    exerciseId := uuid.MustParse(e.ExerciseId)
    sets[i] = model.ExerciseSet{
      ExerciseID: exerciseId,
      SetNumber:  e.SetNumber,
      WeightKg:   float64(e.WeightKg),
      Reps:       e.Reps,
      Failure:    e.Failure,
    }
  }

  resp := &model.Session{
    ProgramDayID: programDayId,
    WeekNumber:   sessionPB.WeekNumber,
    StartedAt:    sessionPB.StartedAt.AsTime(),
    EndedAt:      endedAt,
    Notes:        &sessionPB.Notes,
    ExerciseSets: sets, 
  }

  return resp
}
