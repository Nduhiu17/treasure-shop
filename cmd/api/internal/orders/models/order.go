package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Order struct {
	ID                         primitive.ObjectID  `bson:"_id,omitempty" json:"id,omitempty"`
	UserID                     primitive.ObjectID  `bson:"user_id" json:"user_id"`
	OrderTypeID                primitive.ObjectID  `bson:"order_type_id" json:"order_type_id"` // Foreign key to OrderType
	Title                      string              `bson:"title" json:"title"`
	Description                string              `bson:"description" json:"description"`
	Price                      float64             `bson:"price" json:"price"`
	Status                     string              `bson:"status" json:"status"` // pending_payment, awaiting_assignment, assigned, in_progress, submitted_for_review, approved, feedback, completed
	WriterID                   *primitive.ObjectID `bson:"writer_id,omitempty" json:"writer_id"`
	WriterName                 string              `bson:"-" json:"writer_name,omitempty"`
	SubmissionDate             *time.Time          `bson:"submission_date,omitempty" json:"submission_date,omitempty"`
	Feedback                   string              `bson:"feedback,omitempty" json:"feedback,omitempty"`
	CreatedAt                  time.Time           `bson:"created_at,omitempty" json:"created_at,omitempty"`
	UpdatedAt                  time.Time           `bson:"updated_at,omitempty" json:"updated_at,omitempty"`
	ApplyFeedbackRequests      int                 `bson:"apply_feedback_requests" json:"apply_feedback_requests"`
	OrderLevelID               primitive.ObjectID  `bson:"order_level_id" json:"order_level_id" binding:"required"`
	LevelName                  string              `bson:"-" json:"level_name,omitempty"`
	OrderPagesID               primitive.ObjectID  `bson:"order_pages_id" json:"order_pages_id" binding:"required"`
	OrderPagesName             string              `bson:"-" json:"order_pages_name,omitempty"`
	OrderUrgencyID             primitive.ObjectID  `bson:"order_urgency_id" json:"order_urgency_id" binding:"required"`
	OrderUrgencyName           string              `bson:"-" json:"order_urgency_name,omitempty"`
	IsHighPriority             bool                `bson:"is_high_priority" json:"is_high_priority"`
	OrderStyleID               primitive.ObjectID  `bson:"order_style_id" json:"order_style_id" binding:"required"`
	OrderStyleName             string              `bson:"-" json:"order_style_name,omitempty"`
	OrderLanguageID            primitive.ObjectID  `bson:"order_language_id" json:"order_language_id" binding:"required"`
	OrderLanguageName          string              `bson:"-" json:"order_language_name,omitempty"`
	TopWriter                  bool                `bson:"top_writer" json:"top_writer"`
	PlagarismReport            bool                `bson:"plagarism_report" json:"plagarism_report"`
	OnePageSummary             bool                `bson:"one_page_summary" json:"one_page_summary"`
	ExtraQualityCheck          bool                `bson:"extra_quality_check" json:"extra_quality_check"`
	InitialDraft               bool                `bson:"initial_draft" json:"initial_draft"`
	SmsUpdate                  bool                `bson:"sms_update" json:"sms_update"`
	FullTextCopySources        bool                `bson:"full_text_copy_sources" json:"full_text_copy_sources"`
	SamePaperFromAnotherWriter bool                `bson:"same_paper_from_another_writer" json:"same_paper_from_another_writer"`
	NoOfSources                int                 `bson:"no_of_sources" json:"no_of_sources"`
	PreferredWriterNumber      *int                `bson:"preferred_writer_number,omitempty" json:"preferred_writer_number,omitempty"`
}
