#include <stdlib.h>
#include <assert.h>
#include <string.h>

void f__realloc (char * v__str1, int v__len1, char * v__str2, int v__len2) {  if (v__len1 != v__len2) {
    v__len1 = v__len2;
    v__str1 = (char *) realloc(v__str1, v__len1 * sizeof(char) + 1);
    assert(v__str1 != NULL);
  }
  strcpy(v__str1, v__str2);
}

int v_int = 42;
char *v_str;
int len_v_str;

int main() {
  v_int = 67;
  v_str = strdup("Word");
  f__realloc(v_str, len_v_str, "Other", 5);
  len_v_str = 5;
  free(v_str);
}
