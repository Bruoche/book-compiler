#include <stdlib.h>
#include <string.h>
#include <stdio.h>

void f__realloc (char ** v__str1, int* v__len1, char * v__str2, int v__len2) {
  if (*v__len1 != v__len2) {
    char *_tmp_realloc = (char *) realloc(*v__str1, v__len2 * sizeof(char) + 1);
    if (_tmp_realloc == NULL) {       fprintf(stderr, "ERROR OOM: Out of memory, string could not be resized.\n");       exit(1);     }
    *v__str1 = _tmp_realloc;
    *v__len1 = v__len2;
  }
  memcpy(*v__str1, v__str2, v__len2+1);
}

void f_print (char * v__str) {
  printf("%s\n", v__str);
}

void f_put (char * v__str) {
  printf("%s", v__str);
}

char *v_salutation; // ./test.bk:1
int len_v_salutation = 6; // ./test.bk:1

int main() {
  v_salutation = strdup("Hello "); // ./test.bk:1
  if (v_salutation == NULL) {    fprintf(stderr, "ERROR OOM: Out of memory, string \"Hello \" could not be duplicated into v_salutation.\n");     exit(1);  } // ./test.bk:1
    // ./test.bk:4
    char * concat = strcat(v_salutation, "from C!"); // ./test.bk:4
    f_print(concat); // ./test.bk:4
  free(v_salutation);
}
