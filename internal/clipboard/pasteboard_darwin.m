#import <AppKit/AppKit.h>
#import <Foundation/Foundation.h>
#include <limits.h>
#include <stdlib.h>
#include <string.h>
#include "pasteboard_darwin.h"

static char *clipterm_copy_error(const char *message) {
    size_t len = strlen(message);
    char *copy = malloc(len + 1);
    if (copy == NULL) {
        return NULL;
    }
    memcpy(copy, message, len + 1);
    return copy;
}

static CliptermImageResult clipterm_error(const char *message) {
    CliptermImageResult result;
    result.data = NULL;
    result.len = 0;
    result.err = clipterm_copy_error(message);
    return result;
}

static CliptermFilesResult clipterm_files_error(const char *message) {
    CliptermFilesResult result;
    result.paths = NULL;
    result.count = 0;
    result.err = clipterm_copy_error(message);
    return result;
}

static CliptermImageResult clipterm_data_result(NSData *data) {
    CliptermImageResult result;
    result.data = NULL;
    result.len = 0;
    result.err = NULL;

    if (data == nil || [data length] == 0) {
        return clipterm_error("no image in clipboard");
    }

    if ([data length] > INT_MAX) {
        return clipterm_error("clipboard image is too large");
    }

    void *copy = malloc([data length]);
    if (copy == NULL) {
        return clipterm_error("failed to allocate image buffer");
    }

    memcpy(copy, [data bytes], [data length]);
    result.data = copy;
    result.len = (int)[data length];
    return result;
}

static NSData *clipterm_png_from_tiff(NSData *tiffData) {
    if (tiffData == nil || [tiffData length] == 0) {
        return nil;
    }

    NSBitmapImageRep *rep = [NSBitmapImageRep imageRepWithData:tiffData];
    if (rep == nil) {
        return nil;
    }

    return [rep representationUsingType:NSBitmapImageFileTypePNG properties:@{}];
}

static NSData *clipterm_png_from_image(NSImage *image) {
    if (image == nil) {
        return nil;
    }

    NSData *tiffData = [image TIFFRepresentation];
    return clipterm_png_from_tiff(tiffData);
}

CliptermImageResult clipterm_read_clipboard_image_png(void) {
    @autoreleasepool {
        NSPasteboard *pasteboard = [NSPasteboard generalPasteboard];

        NSData *pngData = [pasteboard dataForType:NSPasteboardTypePNG];
        if (pngData != nil && [pngData length] > 0) {
            return clipterm_data_result(pngData);
        }

        NSData *tiffData = [pasteboard dataForType:NSPasteboardTypeTIFF];
        NSData *convertedPNG = clipterm_png_from_tiff(tiffData);
        if (convertedPNG != nil && [convertedPNG length] > 0) {
            return clipterm_data_result(convertedPNG);
        }

        NSArray *objects = [pasteboard readObjectsForClasses:@[[NSImage class]] options:@{}];
        NSImage *image = [objects firstObject];
        NSData *imagePNG = clipterm_png_from_image(image);
        if (imagePNG != nil && [imagePNG length] > 0) {
            return clipterm_data_result(imagePNG);
        }

        return clipterm_error("no image in clipboard");
    }
}

void clipterm_free_image_result(CliptermImageResult result) {
    if (result.data != NULL) {
        free(result.data);
    }
    if (result.err != NULL) {
        free(result.err);
    }
}

CliptermFilesResult clipterm_read_clipboard_files(void) {
    @autoreleasepool {
        NSPasteboard *pasteboard = [NSPasteboard generalPasteboard];
        NSArray *urls = [pasteboard readObjectsForClasses:@[[NSURL class]] options:@{
            NSPasteboardURLReadingFileURLsOnlyKey: @YES
        }];

        if (urls == nil || [urls count] == 0) {
            return clipterm_files_error("no file in clipboard");
        }

        if ([urls count] > INT_MAX) {
            return clipterm_files_error("too many files in clipboard");
        }

        char **paths = calloc([urls count], sizeof(char *));
        if (paths == NULL) {
            return clipterm_files_error("failed to allocate file list");
        }

        int count = 0;
        for (NSURL *url in urls) {
            if (![url isFileURL]) {
                continue;
            }

            NSString *path = [url path];
            if (path == nil) {
                continue;
            }

            const char *utf8 = [path UTF8String];
            if (utf8 == NULL) {
                continue;
            }

            paths[count] = clipterm_copy_error(utf8);
            if (paths[count] == NULL) {
                for (int i = 0; i < count; i++) {
                    free(paths[i]);
                }
                free(paths);
                return clipterm_files_error("failed to copy file path");
            }
            count++;
        }

        if (count == 0) {
            free(paths);
            return clipterm_files_error("no file in clipboard");
        }

        CliptermFilesResult result;
        result.paths = paths;
        result.count = count;
        result.err = NULL;
        return result;
    }
}

void clipterm_free_files_result(CliptermFilesResult result) {
    if (result.paths != NULL) {
        for (int i = 0; i < result.count; i++) {
            if (result.paths[i] != NULL) {
                free(result.paths[i]);
            }
        }
        free(result.paths);
    }
    if (result.err != NULL) {
        free(result.err);
    }
}

char *clipterm_write_clipboard_text(const char *text) {
    @autoreleasepool {
        if (text == NULL) {
            return clipterm_copy_error("text is null");
        }

        NSString *string = [NSString stringWithUTF8String:text];
        if (string == nil) {
            return clipterm_copy_error("text is not valid UTF-8");
        }

        NSPasteboard *pasteboard = [NSPasteboard generalPasteboard];
        [pasteboard clearContents];
        BOOL ok = [pasteboard setString:string forType:NSPasteboardTypeString];
        if (!ok) {
            return clipterm_copy_error("failed to write clipboard text");
        }

        return NULL;
    }
}

void clipterm_free_error(char *err) {
    if (err != NULL) {
        free(err);
    }
}
