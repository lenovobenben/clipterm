#ifndef CLIPTERM_PASTEBOARD_DARWIN_H
#define CLIPTERM_PASTEBOARD_DARWIN_H

#include <stddef.h>

typedef struct {
    void *data;
    int len;
    char *err;
} CliptermImageResult;

typedef struct {
    char **paths;
    int count;
    char *err;
} CliptermFilesResult;

CliptermImageResult clipterm_read_clipboard_image_png(void);
void clipterm_free_image_result(CliptermImageResult result);
CliptermFilesResult clipterm_read_clipboard_files(void);
void clipterm_free_files_result(CliptermFilesResult result);
char *clipterm_write_clipboard_text(const char *text);
void clipterm_free_error(char *err);

#endif
