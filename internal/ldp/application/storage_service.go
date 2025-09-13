package application

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
)

// FormatConverter defines the interface for RDF format conversion
type FormatConverter interface {
	Convert(data []byte, fromFormat, toFormat string) ([]byte, error)
	ValidateFormat(format string) bool
}

// UnitOfWorkFactory creates new UnitOfWork instances
type UnitOfWorkFactory func() pericarpdomain.UnitOfWork

// StorageService orchestrates storage operations with content negotiation and streaming support
type StorageService struct {
	repo              domain.ResourceRepository
	converter         FormatConverter
	unitOfWorkFactory UnitOfWorkFactory
	mu                sync.RWMutex // For concurrent access handling
}

// NewStorageService creates a new storage service instance
func NewStorageService(
	repo domain.ResourceRepository,
	converter FormatConverter,
	unitOfWorkFactory UnitOfWorkFactory,
) *StorageService {
	return &StorageService{
		repo:              repo,
		converter:         converter,
		unitOfWorkFactory: unitOfWorkFactory,
	}
}

// StoreResource stores a resource with content negotiation support
func (s *StorageService) StoreResource(ctx context.Context, id string, data []byte, contentType string) (*domain.Resource, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate input
	if id == "" {
		return nil, domain.ErrInvalidID.WithOperation("StoreResource")
	}
	if len(data) == 0 {
		return nil, domain.ErrInvalidResource.WithOperation("StoreResource").WithContext("reason", "empty data")
	}

	// Normalize content type
	normalizedContentType := s.normalizeContentType(contentType)

	// Validate format if it's an RDF format
	if s.isRDFFormat(normalizedContentType) && !s.converter.ValidateFormat(normalizedContentType) {
		return nil, domain.ErrUnsupportedFormat.WithOperation("StoreResource").WithContext("format", contentType)
	}

	// Check if resource already exists (for business logic, not for repository updates)
	exists, err := s.repo.Exists(ctx, id)
	if err != nil {
		return nil, domain.WrapStorageError(err, "EXISTENCE_CHECK_FAILED", "failed to check resource existence").WithOperation("StoreResource")
	}

	// Create or update resource (in-memory only)
	var resource *domain.Resource
	if exists {
		// Update existing resource - retrieve current state
		resource, err = s.repo.Retrieve(ctx, id)
		if err != nil {
			return nil, domain.WrapStorageError(err, "RETRIEVE_FAILED", "failed to retrieve existing resource").WithOperation("StoreResource")
		}
		resource.Update(data, normalizedContentType)
	} else {
		// Create new resource
		resource = domain.NewResource(id, normalizedContentType, data)
	}

	// Check if resource is valid before proceeding
	if !resource.IsValid() {
		errors := resource.Errors()
		if len(errors) > 0 {
			return nil, domain.WrapStorageError(errors[0], "INVALID_RESOURCE", "resource validation failed").WithOperation("StoreResource")
		}
		return nil, domain.ErrInvalidResource.WithOperation("StoreResource").WithContext("reason", "resource is not valid")
	}

	// Create a new unit of work for this operation
	unitOfWork := s.unitOfWorkFactory()

	// Register events with unit of work for persistence and dispatch
	events := resource.UncommittedEvents()
	if len(events) > 0 {
		unitOfWork.RegisterEvents(events)
	}

	// Store the resource in repository first (for immediate consistency)
	if err := s.repo.Store(ctx, resource); err != nil {
		// Rollback unit of work on repository failure
		if rollbackErr := unitOfWork.Rollback(); rollbackErr != nil {
			fmt.Printf("Warning: failed to rollback unit of work: %v\n", rollbackErr)
		}
		return nil, domain.WrapStorageError(err, "STORE_FAILED", "failed to store resource").WithOperation("StoreResource")
	}

	// Commit unit of work - this persists events to event store and dispatches them
	// This provides event sourcing capabilities while maintaining immediate consistency
	envelopes, err := unitOfWork.Commit(ctx)
	if err != nil {
		return nil, domain.WrapStorageError(err, "EVENT_COMMIT_FAILED", "failed to commit events").WithOperation("StoreResource")
	}

	// Mark events as committed on the resource
	resource.MarkEventsAsCommitted()

	// Log successful event processing (in production, use proper logger)
	if len(envelopes) > 0 {
		fmt.Printf("Successfully processed %d events for resource %s\n", len(envelopes), id)
	}

	return resource, nil
}

// RetrieveResource retrieves a resource with content negotiation
func (s *StorageService) RetrieveResource(ctx context.Context, id string, acceptFormat string) (*domain.Resource, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Validate input
	if id == "" {
		return nil, domain.ErrInvalidID.WithOperation("RetrieveResource")
	}

	// Retrieve the resource
	resource, err := s.repo.Retrieve(ctx, id)
	if err != nil {
		if domain.IsResourceNotFound(err) {
			return nil, domain.ErrResourceNotFound.WithOperation("RetrieveResource").WithContext("id", id)
		}
		return nil, domain.WrapStorageError(err, "RETRIEVE_FAILED", "failed to retrieve resource").WithOperation("RetrieveResource")
	}

	// Handle content negotiation
	if acceptFormat != "" && acceptFormat != resource.GetContentType() {
		convertedResource, err := s.convertResourceFormat(resource, acceptFormat)
		if err != nil {
			return nil, err
		}
		return convertedResource, nil
	}

	return resource, nil
}

// DeleteResource deletes a resource
func (s *StorageService) DeleteResource(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate input
	if id == "" {
		return domain.ErrInvalidID.WithOperation("DeleteResource")
	}

	// Check if resource exists
	exists, err := s.repo.Exists(ctx, id)
	if err != nil {
		return domain.WrapStorageError(err, "EXISTENCE_CHECK_FAILED", "failed to check resource existence").WithOperation("DeleteResource")
	}
	if !exists {
		return domain.ErrResourceNotFound.WithOperation("DeleteResource").WithContext("id", id)
	}

	// Retrieve resource to emit delete event
	resource, err := s.repo.Retrieve(ctx, id)
	if err != nil {
		return domain.WrapStorageError(err, "RETRIEVE_FAILED", "failed to retrieve resource for deletion").WithOperation("DeleteResource")
	}

	// Mark as deleted (this will add delete event)
	resource.Delete()

	// Create a new unit of work for this operation
	unitOfWork := s.unitOfWorkFactory()

	// Register delete events with unit of work
	events := resource.UncommittedEvents()
	if len(events) > 0 {
		unitOfWork.RegisterEvents(events)
	}

	// Delete from repository first (for immediate consistency)
	if err := s.repo.Delete(ctx, id); err != nil {
		// Rollback unit of work on repository failure
		if rollbackErr := unitOfWork.Rollback(); rollbackErr != nil {
			fmt.Printf("Warning: failed to rollback unit of work: %v\n", rollbackErr)
		}
		return domain.WrapStorageError(err, "DELETE_FAILED", "failed to delete resource").WithOperation("DeleteResource")
	}

	// Commit unit of work - this persists events to event store and dispatches them
	// This provides event sourcing capabilities while maintaining immediate consistency
	envelopes, err := unitOfWork.Commit(ctx)
	if err != nil {
		return domain.WrapStorageError(err, "EVENT_COMMIT_FAILED", "failed to commit delete events").WithOperation("DeleteResource")
	}

	// Mark events as committed on the resource
	resource.MarkEventsAsCommitted()

	// Log successful event processing (in production, use proper logger)
	if len(envelopes) > 0 {
		fmt.Printf("Successfully processed %d delete events for resource %s\n", len(envelopes), id)
	}

	return nil
}

// StreamResource provides streaming access to large resources
func (s *StorageService) StreamResource(ctx context.Context, id string, acceptFormat string) (io.ReadCloser, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// For now, retrieve the full resource and wrap it in a reader
	// In a production implementation, this would stream directly from storage
	resource, err := s.RetrieveResource(ctx, id, acceptFormat)
	if err != nil {
		return nil, "", err
	}

	reader := &resourceReader{
		data:   resource.GetData(),
		offset: 0,
	}

	return reader, resource.GetContentType(), nil
}

// StoreResourceStream stores a resource from a stream
func (s *StorageService) StoreResourceStream(ctx context.Context, id string, reader io.Reader, contentType string) (*domain.Resource, error) {
	// Read all data from stream
	// In a production implementation, this would handle streaming more efficiently
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, domain.WrapStorageError(err, "STREAM_READ_FAILED", "failed to read from stream").WithOperation("StoreResourceStream")
	}

	return s.StoreResource(ctx, id, data, contentType)
}

// ResourceExists checks if a resource exists
func (s *StorageService) ResourceExists(ctx context.Context, id string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if id == "" {
		return false, domain.ErrInvalidID.WithOperation("ResourceExists")
	}

	exists, err := s.repo.Exists(ctx, id)
	if err != nil {
		return false, domain.WrapStorageError(err, "EXISTENCE_CHECK_FAILED", "failed to check resource existence").WithOperation("ResourceExists")
	}

	return exists, nil
}

// convertResourceFormat converts a resource to the requested format
func (s *StorageService) convertResourceFormat(resource *domain.Resource, targetFormat string) (*domain.Resource, error) {
	normalizedTargetFormat := s.normalizeContentType(targetFormat)

	// Validate target format
	if s.isRDFFormat(normalizedTargetFormat) && !s.converter.ValidateFormat(normalizedTargetFormat) {
		return nil, domain.ErrUnsupportedFormat.WithOperation("convertResourceFormat").WithContext("format", targetFormat)
	}

	// If it's not an RDF format, we can't convert
	if !s.isRDFFormat(resource.GetContentType()) || !s.isRDFFormat(normalizedTargetFormat) {
		return nil, domain.ErrFormatConversion.WithOperation("convertResourceFormat").WithContext("reason", "non-RDF format conversion not supported")
	}

	// Convert the data
	convertedData, err := s.converter.Convert(resource.GetData(), resource.GetContentType(), normalizedTargetFormat)
	if err != nil {
		return nil, domain.WrapStorageError(err, "FORMAT_CONVERSION_FAILED", "failed to convert resource format").WithOperation("convertResourceFormat")
	}

	// Create a new resource with converted data
	convertedResource := domain.NewResource(resource.ID(), normalizedTargetFormat, convertedData)

	// Copy metadata from original resource
	for key, value := range resource.GetMetadata() {
		convertedResource.SetMetadata(key, value)
	}
	convertedResource.SetMetadata("convertedFrom", resource.GetContentType())

	return convertedResource, nil
}

// normalizeContentType normalizes content type strings
func (s *StorageService) normalizeContentType(contentType string) string {
	contentType = strings.ToLower(strings.TrimSpace(contentType))

	// Handle common variations
	switch contentType {
	case "json-ld", "jsonld", "application/json":
		return "application/ld+json"
	case "turtle", "ttl", "text/plain":
		return "text/turtle"
	case "rdf/xml", "rdfxml", "xml":
		return "application/rdf+xml"
	default:
		return contentType
	}
}

// isRDFFormat checks if a content type represents an RDF format
func (s *StorageService) isRDFFormat(contentType string) bool {
	rdfFormats := map[string]bool{
		"application/ld+json": true,
		"text/turtle":         true,
		"application/rdf+xml": true,
	}
	return rdfFormats[contentType]
}

// resourceReader implements io.ReadCloser for streaming resource data
type resourceReader struct {
	data   []byte
	offset int
}

func (r *resourceReader) Read(p []byte) (n int, err error) {
	if r.offset >= len(r.data) {
		return 0, io.EOF
	}

	n = copy(p, r.data[r.offset:])
	r.offset += n
	return n, nil
}

func (r *resourceReader) Close() error {
	// Nothing to close for in-memory data
	return nil
}
