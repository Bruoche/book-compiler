#include <stdlib.h>

int v_int = 42;

int main() {
  v_int = 67;
  v_str = (char *) malloc(4 * sizeof(char));
  v_str[0] = 't';
  v_str[1] = 'e';
  v_str[2] = 's';
  v_str[3] = 't';
  free(v_str);
}
