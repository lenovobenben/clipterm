#ifndef CLIPTERM_PASTEBOARD_DARWIN_H
#define CLIPTERM_PASTEBOARD_DARWIN_H

#include <stddef.h>

typedef struct {
    void *data;
    int len;
    char *err;
} CliptermImageResult;

CliptermImageResult clipterm_read_clipboard_image_png(void);
void clipterm_free_image_result(CliptermImageResult result);

#endif
