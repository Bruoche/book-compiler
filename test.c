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

int v_num = 1;
char *v_hi;
int len_v_hi = 5;
char *v_name;
int len_v_name = 5;

int main() {
  v_hi = strdup("hello");
  if (v_hi == NULL) {    fprintf(stderr, "ERROR OOM: Out of memory, string \"hello\" could not be duplicated into v_hi.\n");     exit(1);  }
  f_print(v_hi);
  f__realloc(&v_hi, &len_v_hi, "Hello ", 6);
  v_num = 4;
  f_put(hi);
  v_name = strdup("Alice");
  if (v_name == NULL) {    fprintf(stderr, "ERROR OOM: Out of memory, string \"Alice\" could not be duplicated into v_name.\n");     exit(1);  }
  f_print(v_name);
  f__realloc(&v_hi, &len_v_hi, "... ", 4);
  f_put(v_hi);
  f_print(v_name);
  free(v_hi);
  free(v_name);
}
