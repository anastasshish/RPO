/*-
 * Read/write MIFARE Classic 1K sector 1 data blocks (4, 5, 6) only.
 * Avoids authenticating sectors 2+ when their keys differ from default.
 *
 * Usage:
 *   nfc-mfsector1 r <out.bin>   - write 48 bytes (blocks 4-6) to file
 *   nfc-mfsector1 w <in.bin>    - read 48 bytes from file and write to card
 */

#ifdef HAVE_CONFIG_H
#  include "config.h"
#endif

#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>
#include <string.h>

#ifndef _WIN32
#  include <unistd.h>
#endif

#include <nfc/nfc.h>
#include "mifare.h"

static const nfc_modulation nmMifare = {
  .nmt = NMT_ISO14443A,
  .nbr = NBR_106,
};

static const uint8_t default_key[] = { 0xff, 0xff, 0xff, 0xff, 0xff, 0xff };
static const uint8_t keys[][6] = {
  { 0xff, 0xff, 0xff, 0xff, 0xff, 0xff },
  { 0xd3, 0xf7, 0xd3, 0xf7, 0xd3, 0xf7 },
  { 0xa0, 0xa1, 0xa2, 0xa3, 0xa4, 0xa5 },
  { 0xb0, 0xb1, 0xb2, 0xb3, 0xb4, 0xb5 },
  { 0x4d, 0x3a, 0x99, 0xc3, 0x51, 0xdd },
  { 0x1a, 0x98, 0x2c, 0x7e, 0x45, 0x9a },
  { 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff },
  { 0x00, 0x00, 0x00, 0x00, 0x00, 0x00 },
  { 0xab, 0xcd, 0xef, 0x12, 0x34, 0x56 },
};

static bool
authenticate_sector1(nfc_device *pnd, nfc_target *nt)
{
  mifare_param mp;
  size_t i;

  memcpy(mp.mpa.abtAuthUid, nt->nti.nai.abtUid + nt->nti.nai.szUidLen - 4, 4);

  for (i = 0; i < sizeof(keys) / sizeof(keys[0]); i++) {
    memcpy(mp.mpa.abtKey, keys[i], 6);
    if (nfc_initiator_mifare_cmd(pnd, MC_AUTH_A, 4, &mp)) {
      return true;
    }
    if (nfc_initiator_select_passive_target(pnd, nmMifare, nt->nti.nai.abtUid,
                                            nt->nti.nai.szUidLen, NULL) <= 0) {
      return false;
    }
  }

  memcpy(mp.mpa.abtKey, default_key, 6);
  return nfc_initiator_mifare_cmd(pnd, MC_AUTH_A, 4, &mp);
}

static bool
read_block(nfc_device *pnd, uint8_t block, uint8_t *out16)
{
  mifare_param mp;

  if (!nfc_initiator_mifare_cmd(pnd, MC_READ, block, &mp)) {
    return false;
  }
  memcpy(out16, mp.mpd.abtData, 16);
  return true;
}

static bool
write_block(nfc_device *pnd, uint8_t block, const uint8_t *in16)
{
  mifare_param mp;

  memcpy(mp.mpd.abtData, in16, 16);
  return nfc_initiator_mifare_cmd(pnd, MC_WRITE, block, &mp);
}

static void
usage(const char *name)
{
  fprintf(stderr,
          "Usage: %s r <out.bin>  |  %s w <in.bin>\n"
          "  r - read blocks 4,5,6 (48 bytes) from card\n"
          "  w - write 48 bytes to blocks 4,5,6\n",
          name, name);
}

int
main(int argc, char *argv[])
{
  nfc_context *context;
  nfc_device *pnd;
  nfc_target nt;
  uint8_t data[48];
  FILE *f;
  size_t n;
  bool write_mode;
  int block;

  if (argc != 3) {
    usage(argv[0]);
    return EXIT_FAILURE;
  }

  write_mode = (argv[1][0] == 'w');
  if (!write_mode && argv[1][0] != 'r') {
    usage(argv[0]);
    return EXIT_FAILURE;
  }

  if (write_mode) {
    f = fopen(argv[2], "rb");
    if (!f) {
      perror("fopen");
      return EXIT_FAILURE;
    }
    n = fread(data, 1, sizeof(data), f);
    fclose(f);
    if (n != sizeof(data)) {
      fprintf(stderr, "Need exactly 48 bytes in %s\n", argv[2]);
      return EXIT_FAILURE;
    }
  }

  nfc_init(&context);
  pnd = nfc_open(context, NULL);
  if (!pnd) {
    fprintf(stderr, "Unable to open NFC device\n");
    nfc_exit(context);
    return EXIT_FAILURE;
  }

  if (nfc_initiator_init(pnd) < 0) {
    nfc_perror(pnd, "nfc_initiator_init");
    nfc_close(pnd);
    nfc_exit(context);
    return EXIT_FAILURE;
  }

  if (nfc_initiator_select_passive_target(pnd, nmMifare, NULL, 0, &nt) <= 0) {
    fprintf(stderr, "No tag found\n");
    nfc_close(pnd);
    nfc_exit(context);
    return EXIT_FAILURE;
  }

  if (!authenticate_sector1(pnd, &nt)) {
    fprintf(stderr, "Authentication failed for sector 1 (blocks 4-7)\n");
    nfc_close(pnd);
    nfc_exit(context);
    return EXIT_FAILURE;
  }

  if (write_mode) {
    for (block = 0; block < 3; block++) {
      if (!write_block(pnd, (uint8_t)(4 + block), data + block * 16)) {
        fprintf(stderr, "Write failed for block %d\n", 4 + block);
        nfc_close(pnd);
        nfc_exit(context);
        return EXIT_FAILURE;
      }
    }
    printf("Wrote blocks 4, 5, 6\n");
  } else {
    for (block = 0; block < 3; block++) {
      if (!read_block(pnd, (uint8_t)(4 + block), data + block * 16)) {
        fprintf(stderr, "Read failed for block %d\n", 4 + block);
        nfc_close(pnd);
        nfc_exit(context);
        return EXIT_FAILURE;
      }
    }
    f = fopen(argv[2], "wb");
    if (!f) {
      perror("fopen");
      nfc_close(pnd);
      nfc_exit(context);
      return EXIT_FAILURE;
    }
    if (fwrite(data, 1, sizeof(data), f) != sizeof(data)) {
      fprintf(stderr, "Could not write %s\n", argv[2]);
      fclose(f);
      nfc_close(pnd);
      nfc_exit(context);
      return EXIT_FAILURE;
    }
    fclose(f);
    printf("Read blocks 4, 5, 6 -> %s\n", argv[2]);
  }

  nfc_close(pnd);
  nfc_exit(context);
  return EXIT_SUCCESS;
}
