package processor

import "github.com/SirNacou/refract/services/analytics-processor/internal/domain"
type BatchData struct {
    events []domain.ClickEvent
    ids    []string
}