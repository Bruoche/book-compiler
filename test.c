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

int v_int = 42;
char *v_salutation;
int len_v_salutation = 5;
char *v_word;
int len_v_word = 6;

int main() {
  v_int = 67;
  v_salutation = strdup("Hello");
  if (v_salutation == NULL) {    fprintf(stderr, "ERROR OOM: Out of memory, string \"Hello\" could not be duplicated into v_salutation.\n");     exit(1);  }
  f__realloc(&v_salutation, &len_v_salutation, "Hello ", 6);
  f_put(v_salutation);
  v_word = strdup("Word!!");
  if (v_word == NULL) {    fprintf(stderr, "ERROR OOM: Out of memory, string \"Word!!\" could not be duplicated into v_word.\n");     exit(1);  }
  f_print(v_word);
  free(v_salutation);
  free(v_word);
}
