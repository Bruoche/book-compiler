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

int v_num = 1; // ./test.bk:1
char *v_hi; // ./test.bk:2
int len_v_hi = 5; // ./test.bk:2
char *v_name; // ./test.bk:7
int len_v_name = 5; // ./test.bk:7

int main() {
  v_hi = strdup("hello"); // ./test.bk:2
  if (v_hi == NULL) {    fprintf(stderr, "ERROR OOM: Out of memory, string \"hello\" could not be duplicated into v_hi.\n");     exit(1);  } // ./test.bk:2
  f_print(v_hi); // ./test.bk:3
  f__realloc(&v_hi, &len_v_hi, "Hello ", 6); // ./test.bk:4
  v_num = 4; // ./test.bk:5
  f_put(hi); // ./test.bk:6
  v_name = strdup("Alice"); // ./test.bk:7
  if (v_name == NULL) {    fprintf(stderr, "ERROR OOM: Out of memory, string \"Alice\" could not be duplicated into v_name.\n");     exit(1);  } // ./test.bk:7
  f_print(v_name); // ./test.bk:8
  f__realloc(&v_hi, &len_v_hi, "... ", 4); // ./test.bk:9
  f_put(v_hi); // ./test.bk:10
  f_print(v_name); // ./test.bk:10
  free(v_hi);
  free(v_name);
}
