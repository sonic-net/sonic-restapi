

nobase_include_HEADERS = sonic/db_sql_defines.h  sonic/db_sql_ops.h
#The sdi-sys library
lib_LTLIBRARIES = libsonic_db_sql.la

libsonic_db_sql_la_SOURCES = src/db_sql_ops.c


libsonic_db_sql_la_LDFLAGS = -lsqlite3  -lsonic_common -lsonic_logging -lpthread -lrt -version-info 1:1:0
libsonic_db_sql__la_CFLAGS = -I$(top_srcdir)/sonic  -I$(includedir)/sonic  -I$(top_srcdir)/sonic/private


