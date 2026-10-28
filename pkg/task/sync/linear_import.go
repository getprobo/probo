// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package tasksync

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/url"
	"path"
	"sort"
	"strings"
	"sync"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/filevalidation"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/prosemirror"
	"go.probo.inc/probo/pkg/task/sync/linear"
)

type importedFileMeta struct {
	Name     string
	MimeType string
	Size     int64
}

// linearImport is markdown with Linear uploads already copied into Probo.
type linearImport struct {
	connectorID      gid.GID
	organizationID   gid.GID
	markdown         string
	files            map[string]string
	metas            map[string]importedFileMeta
	created          []createdInboundFile
	mentionsApplied  bool
	mentionContent   string
	mentionCanonical string
}

// createdInboundFile is a private file stored by this import. A discarded
// event deletes it, because nothing else references the row yet.
type createdInboundFile struct {
	id          gid.GID
	connectorID gid.GID
	assetURL    string
}

// inboundFileCache remembers a Linear upload stored during this process.
// The file row commits on its own connection, so a rolled-back task import
// would otherwise download and store the same upload again. One in-flight
// import per connector and URL keeps concurrent webhooks from storing it twice.
type inboundFileCache struct {
	mu       sync.Mutex
	files    map[string]cachedInboundFile
	inflight map[string]*inboundFlight
}

type inboundFlight struct {
	done chan struct{}
	file cachedInboundFile
	err  error
}

type cachedInboundFile struct {
	fileID string
	meta   importedFileMeta
}

// importLinearMarkdown downloads Linear uploads referenced by markdown and
// rewrites them to stored file paths. known maps a file id to an asset URL
// already stored for this task. Images become image nodes. Other files become
// file nodes when the link is the whole paragraph.
func (s *Service) importLinearMarkdown(
	ctx context.Context,
	client *linear.Client,
	scope coredata.Scoper,
	connectorID gid.GID,
	organizationID gid.GID,
	markdown string,
	known map[string]string,
) (string, map[string]string, map[string]importedFileMeta, error) {
	markdown, files, metas, _, err := s.importLinearFiles(
		ctx,
		client,
		scope,
		connectorID,
		organizationID,
		markdown,
		known,
		false,
	)

	return markdown, files, metas, err
}

// importLinearMarkdownBestEffort keeps going when one upload cannot be
// stored. The task or comment is still saved, and the Linear URL stays in
// the text so a later event can try that file again.
func (s *Service) importLinearMarkdownBestEffort(
	ctx context.Context,
	client *linear.Client,
	scope coredata.Scoper,
	connectorID gid.GID,
	organizationID gid.GID,
	markdown string,
	known map[string]string,
) (string, map[string]string, map[string]importedFileMeta, error) {
	markdown, files, metas, _, err := s.importLinearFiles(
		ctx,
		client,
		scope,
		connectorID,
		organizationID,
		markdown,
		known,
		true,
	)

	return markdown, files, metas, err
}

func (s *Service) importLinearFiles(
	ctx context.Context,
	client *linear.Client,
	scope coredata.Scoper,
	connectorID gid.GID,
	organizationID gid.GID,
	markdown string,
	known map[string]string,
	continueOnFailure bool,
) (string, map[string]string, map[string]importedFileMeta, []createdInboundFile, error) {
	s.forgetKnownInboundFiles(connectorID, known)
	markdown = restoreLinearFiles(markdown, known)

	urls := linearAssetURLs(markdown)
	if len(urls) == 0 {
		metas, err := s.knownFileMetas(ctx, scope, organizationID, referencedAttachments(markdown, known))
		if err != nil {
			return "", nil, nil, nil, err
		}

		return markdown, known, metas, nil, nil
	}

	if s.stores == nil {
		return "", nil, nil, nil, fmt.Errorf("cannot store Linear attachment: storage is not configured")
	}

	if client == nil {
		loaded, err := s.linearClientForConnectorID(ctx, scope, connectorID)
		if err != nil {
			return "", nil, nil, nil, err
		}

		client = loaded
	}

	files := make(map[string]string, len(known)+len(urls))
	maps.Copy(files, known)

	metas := map[string]importedFileMeta{}

	var (
		imagePaths []string
		created    []createdInboundFile
	)

	for _, assetURL := range urls {
		name := linearAssetLabel(markdown, assetURL)
		if name == "" {
			name = linearAssetFileName(assetURL)
		}

		fileID, meta, stored, err := s.resolveInboundFile(
			ctx,
			client,
			scope,
			connectorID,
			organizationID,
			assetURL,
			name,
		)
		if err != nil {
			if continueOnFailure || permanentLinearAsset(err) || permanentStoreError(err) {
				s.warnLinearAttachment(ctx, "inbound", err)

				continue
			}

			return "", nil, nil, nil, err
		}

		if fileID == "" {
			continue
		}

		files[fileID] = assetURL
		metas[fileID] = meta

		if stored {
			parsed, parseErr := gid.ParseGID(fileID)
			if parseErr == nil {
				created = append(created, createdInboundFile{
					id:          parsed,
					connectorID: connectorID,
					assetURL:    assetURL,
				})
			}
		}

		if linearImageType(meta.MimeType) {
			imagePaths = append(imagePaths, prosemirror.AttachmentPath(fileID))
		}
	}

	if err := s.fillKnownFileMetas(
		ctx,
		scope,
		organizationID,
		referencedAttachments(markdown, files),
		metas,
	); err != nil {
		return "", nil, nil, nil, err
	}

	markdown = replaceLinearAssets(markdown, files)
	markdown = promoteStandaloneImages(markdown, imagePaths)

	return markdown, files, metas, created, nil
}

func (s *Service) prepareLinearImport(
	ctx context.Context,
	scope coredata.Scoper,
	connectorID gid.GID,
	organizationID gid.GID,
	markdown string,
	known map[string]string,
	continueOnFailure bool,
) (*linearImport, error) {
	var client *linear.Client

	if len(linearAssetURLs(restoreLinearFiles(markdown, known))) > 0 {
		loaded, err := s.linearClientForConnectorID(ctx, scope, connectorID)
		if err != nil {
			return nil, err
		}

		client = loaded
	}

	rewritten, files, metas, created, err := s.importLinearFiles(
		ctx,
		client,
		scope,
		connectorID,
		organizationID,
		markdown,
		known,
		continueOnFailure,
	)
	if err != nil {
		return nil, err
	}

	imported := &linearImport{
		connectorID:    connectorID,
		organizationID: organizationID,
		markdown:       rewritten,
		files:          files,
		metas:          metas,
		created:        created,
	}
	if err := s.attachLinearMentions(ctx, client, scope, organizationID, imported); err != nil {
		return nil, err
	}

	return imported, nil
}

func (s *Service) resolveInboundFile(
	ctx context.Context,
	client *linear.Client,
	scope coredata.Scoper,
	connectorID gid.GID,
	organizationID gid.GID,
	assetURL string,
	name string,
) (fileID string, meta importedFileMeta, created bool, err error) {
	cached, flight, leader := s.startInboundFlight(connectorID, assetURL)
	if !leader && flight == nil {
		live, ok, liveErr := s.liveInboundFile(ctx, scope, organizationID, cached)
		if liveErr != nil {
			return "", importedFileMeta{}, false, liveErr
		}

		if ok {
			return cached.fileID, live, false, nil
		}

		s.forgetInboundFile(connectorID, assetURL)
		cached, flight, leader = s.startInboundFlight(connectorID, assetURL)
	}

	if !leader {
		if flight == nil {
			return cached.fileID, cached.meta, false, nil
		}

		select {
		case <-ctx.Done():
			return "", importedFileMeta{}, false, ctx.Err()
		case <-flight.done:
		}

		if flight.err != nil {
			return "", importedFileMeta{}, false, flight.err
		}

		return flight.file.fileID, flight.file.meta, false, nil
	}

	defer func() {
		s.finishInboundFlight(inboundFileKey(connectorID, assetURL), flight, cachedInboundFile{
			fileID: fileID,
			meta:   meta,
		}, err)
	}()

	body, contentType, err := client.DownloadAsset(ctx, assetURL)
	if err != nil {
		return "", importedFileMeta{}, false, err
	}

	stored, err := s.stores.StoreAttachment(
		ctx,
		scope,
		organizationID,
		Attachment{
			Name:     name,
			MimeType: contentType,
			Body:     body,
		},
	)
	if err != nil {
		return "", importedFileMeta{}, false, err
	}

	return stored.ID.String(), importedFileMeta{
		Name:     stored.Name,
		MimeType: stored.MimeType,
		Size:     stored.Size,
	}, true, nil
}

func (s *Service) startInboundFlight(
	connectorID gid.GID,
	assetURL string,
) (cachedInboundFile, *inboundFlight, bool) {
	if s == nil || assetURL == "" {
		return cachedInboundFile{}, nil, true
	}

	key := inboundFileKey(connectorID, assetURL)

	s.inboundFiles.mu.Lock()
	defer s.inboundFiles.mu.Unlock()

	if cached, ok := s.inboundFiles.files[key]; ok && cached.fileID != "" {
		return cached, nil, false
	}

	if flight, ok := s.inboundFiles.inflight[key]; ok {
		return cachedInboundFile{}, flight, false
	}

	flight := &inboundFlight{done: make(chan struct{})}

	if s.inboundFiles.inflight == nil {
		s.inboundFiles.inflight = map[string]*inboundFlight{}
	}

	s.inboundFiles.inflight[key] = flight

	return cachedInboundFile{}, flight, true
}

func (s *Service) finishInboundFlight(
	key string,
	flight *inboundFlight,
	file cachedInboundFile,
	err error,
) {
	if s == nil || flight == nil {
		return
	}

	s.inboundFiles.mu.Lock()
	defer s.inboundFiles.mu.Unlock()

	if err == nil && file.fileID != "" {
		if s.inboundFiles.files == nil {
			s.inboundFiles.files = map[string]cachedInboundFile{}
		}

		s.inboundFiles.files[key] = file
		flight.file = file
	}

	flight.err = err

	delete(s.inboundFiles.inflight, key)
	close(flight.done)
}

func (s *Service) liveInboundFile(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	cached cachedInboundFile,
) (importedFileMeta, bool, error) {
	if s == nil || s.pg == nil || scope == nil {
		return cached.meta, true, nil
	}

	fileID, err := gid.ParseGID(cached.fileID)
	if err != nil {
		return importedFileMeta{}, false, nil
	}

	record := &coredata.File{}

	err = s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := record.LoadActiveByID(ctx, conn, scope, fileID); err != nil {
				return err
			}

			return nil
		},
	)
	if errors.Is(err, coredata.ErrResourceNotFound) {
		return importedFileMeta{}, false, nil
	}

	if err != nil {
		return importedFileMeta{}, false, fmt.Errorf("cannot load stored attachment: %w", err)
	}

	if record.OrganizationID != organizationID {
		return importedFileMeta{}, false, nil
	}

	return importedFileMeta{
		Name:     record.FileName,
		MimeType: record.MimeType,
		Size:     record.FileSize,
	}, true, nil
}

func (s *Service) forgetInboundFile(connectorID gid.GID, assetURL string) {
	if s == nil || assetURL == "" {
		return
	}

	s.inboundFiles.mu.Lock()
	defer s.inboundFiles.mu.Unlock()

	delete(s.inboundFiles.files, inboundFileKey(connectorID, assetURL))
}

func (s *Service) forgetKnownInboundFiles(connectorID gid.GID, known map[string]string) {
	for _, assetURL := range known {
		s.forgetInboundFile(connectorID, assetURL)
	}
}

func (s *Service) discardCreatedFiles(
	ctx context.Context,
	scope coredata.Scoper,
	created []createdInboundFile,
) error {
	if s == nil || s.pg == nil || scope == nil || len(created) == 0 {
		return nil
	}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			for _, file := range created {
				record := coredata.File{ID: file.id}
				if err := record.SoftDelete(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot delete unused Linear attachment: %w", err)
				}
			}

			return nil
		},
	)
	if err != nil {
		return err
	}

	for _, file := range created {
		s.forgetInboundFile(file.connectorID, file.assetURL)
	}

	return nil
}

func (s *Service) discardInboundImport(ctx context.Context, imported *linearImport) error {
	if imported == nil {
		return nil
	}

	return s.discardCreatedFiles(
		ctx,
		coredata.NewScope(imported.organizationID.TenantID(), gid.Nil),
		imported.created,
	)
}

func (s *Service) discardPreparedIssues(ctx context.Context, issues []preparedImportedIssue) error {
	for i := range issues {
		if issues[i].connector == nil {
			continue
		}

		err := s.discardCreatedFiles(
			ctx,
			coredata.NewScope(issues[i].connector.OrganizationID.TenantID(), gid.Nil),
			issues[i].created,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func mergeAttachmentMaps(base, extra map[string]string) map[string]string {
	if len(extra) == 0 {
		return base
	}

	merged := make(map[string]string, len(base)+len(extra))
	maps.Copy(merged, base)
	maps.Copy(merged, extra)

	return merged
}

func inboundFileKey(connectorID gid.GID, assetURL string) string {
	return connectorID.String() + "\x00" + assetURL
}

func referencedAttachments(markdown string, files map[string]string) map[string]string {
	if markdown == "" || len(files) == 0 {
		return nil
	}

	referenced := make(map[string]string, len(files))

	for id, assetURL := range files {
		filePath := prosemirror.AttachmentPath(id)
		if filePath != "" && strings.Contains(markdown, filePath) {
			referenced[id] = assetURL

			continue
		}

		if assetURL != "" && strings.Contains(markdown, assetURL) {
			referenced[id] = assetURL
		}
	}

	if len(referenced) == 0 {
		return nil
	}

	return referenced
}

func (s *Service) knownFileMetas(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	files map[string]string,
) (map[string]importedFileMeta, error) {
	metas := map[string]importedFileMeta{}
	if err := s.fillKnownFileMetas(ctx, scope, organizationID, files, metas); err != nil {
		return nil, err
	}

	if len(metas) == 0 {
		return nil, nil
	}

	return metas, nil
}

func (s *Service) fillKnownFileMetas(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	files map[string]string,
	metas map[string]importedFileMeta,
) error {
	if s == nil || s.pg == nil || scope == nil || len(files) == 0 || metas == nil {
		return nil
	}

	type storedID struct {
		raw    string
		parsed gid.GID
	}

	missing := make([]storedID, 0, len(files))

	for id := range files {
		if _, ok := metas[id]; ok {
			continue
		}

		parsed, err := gid.ParseGID(id)
		if err != nil {
			continue
		}

		missing = append(missing, storedID{raw: id, parsed: parsed})
	}

	if len(missing) == 0 {
		return nil
	}

	return s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			for _, item := range missing {
				record := &coredata.File{}

				err := record.LoadByID(ctx, conn, scope, item.parsed)
				if errors.Is(err, coredata.ErrResourceNotFound) {
					continue
				}

				if err != nil {
					return fmt.Errorf("cannot load stored attachment: %w", err)
				}

				if record.OrganizationID != organizationID {
					continue
				}

				metas[item.raw] = importedFileMeta{
					Name:     record.FileName,
					MimeType: record.MimeType,
					Size:     record.FileSize,
				}
			}

			return nil
		},
	)
}

func (s *Service) linearClientForConnectorID(
	ctx context.Context,
	scope coredata.Scoper,
	connectorID gid.GID,
) (*linear.Client, error) {
	if s.pg == nil {
		return nil, fmt.Errorf("cannot load Linear connector: database is not configured")
	}

	dbConnector := &coredata.Connector{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := dbConnector.LoadByID(ctx, conn, scope, connectorID, s.encryptionKey); err != nil {
				return fmt.Errorf("cannot load connector: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load Linear connector: %w", err)
	}

	client, _, err := s.linearClientForConnector(ctx, scope, dbConnector)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func permanentLinearAsset(err error) bool {
	if _, ok := errors.AsType[*linear.AssetUnavailableError](err); ok {
		return true
	}

	_, ok := errors.AsType[*linear.AssetRejectedError](err)

	return ok
}

func permanentStoreError(err error) bool {
	_, ok := errors.AsType[*filevalidation.Error](err)

	return ok
}

func linearImageType(mimeType string) bool {
	switch mimeType {
	case "image/jpeg", "image/png", "image/webp", "image/svg+xml", "image/gif":
		return true
	default:
		return false
	}
}

func linearAssetURLs(markdown string) []string {
	const marker = "https://uploads.linear.app/"

	var urls []string

	seen := map[string]struct{}{}
	rest := markdown

	for {
		index := strings.Index(rest, marker)
		if index < 0 {
			return urls
		}

		raw := rest[index:]
		end := 0

		for end < len(raw) {
			char := raw[end]
			if char <= ' ' || strings.ContainsRune(")>\"'],", rune(char)) {
				break
			}

			end++
		}

		candidate := raw[:end]
		rest = raw[end:]

		if _, ok := seen[candidate]; ok || !linearUploadURL(candidate) {
			continue
		}

		seen[candidate] = struct{}{}
		urls = append(urls, candidate)
	}
}

func linearUploadURL(raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.User != nil {
		return false
	}

	if parsed.Scheme != "https" || !strings.EqualFold(parsed.Hostname(), "uploads.linear.app") {
		return false
	}

	return parsed.Path != "" && parsed.Path != "/"
}

func linearAssetLabel(markdown, assetURL string) string {
	needle := "](" + assetURL + ")"
	index := strings.Index(markdown, needle)

	if index < 0 {
		return ""
	}

	start := strings.LastIndex(markdown[:index], "[")
	if start < 0 || start+1 >= index {
		return ""
	}

	label := strings.TrimSpace(markdown[start+1 : index])
	if label == "" || strings.ContainsAny(label, "\n[]") {
		return ""
	}

	return label
}

func linearAssetFileName(assetURL string) string {
	parsed, err := url.Parse(assetURL)
	if err != nil {
		return "file"
	}

	name := path.Base(parsed.Path)
	if name == "." || name == "/" || name == "" {
		return "file"
	}

	return name
}

func replaceLinearAssets(markdown string, files map[string]string) string {
	if len(files) == 0 || markdown == "" {
		return markdown
	}

	type pair struct {
		asset string
		path  string
	}

	pairs := make([]pair, 0, len(files))

	for id, assetURL := range files {
		if id == "" || assetURL == "" {
			continue
		}

		pairs = append(pairs, pair{asset: assetURL, path: prosemirror.AttachmentPath(id)})
	}

	sort.Slice(pairs, func(i, j int) bool {
		return len(pairs[i].asset) > len(pairs[j].asset)
	})

	for _, item := range pairs {
		markdown = strings.ReplaceAll(markdown, item.asset, item.path)
	}

	return markdown
}

func promoteStandaloneImages(markdown string, paths []string) string {
	if len(paths) == 0 || markdown == "" {
		return markdown
	}

	lines := strings.Split(markdown, "\n")

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "![") {
			continue
		}

		for _, filePath := range paths {
			suffix := "](" + filePath + ")"
			if !strings.HasPrefix(trimmed, "[") || !strings.HasSuffix(trimmed, suffix) || strings.Count(trimmed, "](") != 1 {
				continue
			}

			index := strings.Index(line, "[")
			if index >= 0 {
				lines[i] = line[:index] + "!" + line[index:]
			}

			break
		}
	}

	return strings.Join(lines, "\n")
}

func annotateStoredFiles(content string, metas map[string]importedFileMeta) (string, error) {
	if len(metas) == 0 || strings.TrimSpace(content) == "" {
		return content, nil
	}

	node, err := prosemirror.Parse(content)
	if err != nil {
		return "", fmt.Errorf("cannot parse imported content: %w", err)
	}

	if !annotateFileNodes(&node, metas) {
		return content, nil
	}

	encoded, err := json.Marshal(node)
	if err != nil {
		return "", fmt.Errorf("cannot marshal imported content: %w", err)
	}

	raw := string(encoded)

	sanitized, err := prosemirror.DefaultDocumentJSON(&raw)
	if err != nil {
		return "", fmt.Errorf("cannot sanitize imported content: %w", err)
	}

	return sanitized, nil
}

func annotateFileNodes(node *prosemirror.Node, metas map[string]importedFileMeta) bool {
	changed := false

	if node.Type == prosemirror.NodeFile {
		attrs, err := node.FileAttrs()
		if err == nil {
			if meta, ok := metas[attrs.FileID]; ok {
				if meta.MimeType != "" {
					attrs.MimeType = meta.MimeType
				}

				if meta.Size > 0 {
					attrs.Size = meta.Size
				}

				if strings.TrimSpace(attrs.FileName) == "" && meta.Name != "" {
					attrs.FileName = meta.Name
				}

				raw, marshalErr := json.Marshal(attrs)
				if marshalErr == nil {
					node.Attrs = raw
					changed = true
				}
			}
		}
	}

	for i := range node.Content {
		if annotateFileNodes(&node.Content[i], metas) {
			changed = true
		}
	}

	return changed
}

func applyImportedFiles(content string, metas map[string]importedFileMeta) (string, string, error) {
	if len(metas) == 0 {
		canonical, err := ContentToMarkdown(content)
		if err != nil {
			return "", "", err
		}

		return content, canonical, nil
	}

	annotated, err := annotateStoredFiles(content, metas)
	if err != nil {
		return "", "", err
	}

	canonical, err := ContentToMarkdown(annotated)
	if err != nil {
		return "", "", err
	}

	return annotated, canonical, nil
}

func mergedLinearAttachments(metadata json.RawMessage, files map[string]string) (json.RawMessage, bool, error) {
	if !attachmentMapChanged(attachmentsFromMetadata(metadata), files) {
		return metadata, false, nil
	}

	next, err := mergeAttachments(metadata, files)
	if err != nil {
		return nil, false, fmt.Errorf("cannot merge Linear attachments: %w", err)
	}

	return next, true, nil
}

func attachmentMapChanged(before, after map[string]string) bool {
	for id, assetURL := range after {
		if before[id] != assetURL {
			return true
		}
	}

	return false
}
