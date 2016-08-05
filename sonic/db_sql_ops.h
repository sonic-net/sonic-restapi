/*
 * Copyright (c) 2016 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may
 * not use this file except in compliance with the License. You may obtain
 * a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
 *
 * THIS CODE IS PROVIDED ON AN  *AS IS* BASIS, WITHOUT WARRANTIES OR
 * CONDITIONS OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING WITHOUT
 * LIMITATION ANY IMPLIED WARRANTIES OR CONDITIONS OF TITLE, FITNESS
 * FOR A PARTICULAR PURPOSE, MERCHANTABLITY OR NON-INFRINGEMENT.
 *
 * See the Apache Version 2.0 License for specific language governing
 * permissions and limitations under the License.
 */


/**
 * \file db_sql_ops.h
 * \brief   *Virtual Machine Environment Constant Definitionss
 */

#ifndef _DB_SQL_OPS_H_
#define _DB_SQL_OPS_H_

#include "sonic/std_error_codes.h"

typedef void * db_sql_handle_t;

/** Maximum Length of SQL Command  */
#define DB_SQL_CMD_MAX_LEN 255
/** Database operation code length */
#define DB_SQL_OPCODE_LENGTH 16
/** Maximum number of Database Records */
#define DB_SQL_MAX_RECORDS  4
/** Database query output length */
#define DB_SQL_OUTPUT_LEN 48
/** Maximum length of error string */
#define DB_SQL_MAX_ERROR_LENGTH 64

typedef enum db_sql_opcode_e {
/** CREATE operation  */
    SQL_CREATE,
/** DROP operation  */
    SQL_DROP,
/** INSERT operation  */
    SQL_INSERT,
/** DELETE operation  */
    SQL_DELETE,
/** UPDATE operation  */
    SQL_UPDATE,
/** SELECT operation  */
    SQL_SELECT,
/** TRIGGER operation  */
    SQL_TRIGGER
} db_sql_opcode_t;

typedef enum db_error_code_e {

/** SQL_PREPARE API fail */
    DB_ERR_SQL_PREPARE_FAIL,
/** SQL_STEP API fail */
    DB_ERR_SQL_STEP_FAIL,
/** DB_OPEN API fail */
    DB_ERR_SQL_DB_OPEN_FAIL,
/** DB_CLOSE API fail */
    DB_ERR_SQL_DB_CLOSE_FAIL,
/** DB Operation currently not supported */
    DB_ERR_UNSUPP_OP_TYPE,
/** construct_sql_cmd failure */
    DB_ERR_CONSTRUCT_SQL_CMD
} db_error_code_t;

#ifdef __cplusplus
extern "C" {
#endif

/** \defgroup SQLiteDatabaseOperation DS - SQLite Database Operations
 *  Function Prototype definitions for database operations
 *
 *  \{
 */

/**
 * \brief   Peform database open operation
 * \param   [in] db_name type character string
 * \param   [out] ret_handle type ponter to db_sql_handle_t
 * \return  t_std_error
 */

t_std_error db_sql_open(db_sql_handle_t *ret_handle, const char *db_name);

/**
 * \brief   Perofrm database close operation
 * \param   [in] db_handle type db_sql_handle_t
 * \return  t_std_error
 */

t_std_error db_sql_close(db_sql_handle_t db_handle);

/**
 * \brief   Create the specific table specified by table_name and schema
 * \param   [in] db_handle type db_sql_handle_t
 * \param   [in] table_name type character string
 * \param   [in] schema type character string
 * \return  t_std_error
 */

t_std_error db_sql_create_table(db_sql_handle_t db_handle,
    const char *table_name, const char *schema);

/**
 * \brief   Drop the database table specified by the table_name
 * \param   [in] db_handle type db_sql_handle_t
 * \param   [in] table_name type character string
 * \return  t_std_error
 */

t_std_error db_sql_drop_table(db_sql_handle_t db_handle, const char *table_name);

/**
 * \brief   Insert a record into the database table
 * \param   [in] db_handle type db_sql_handle_t
 * \param   [in] table_name character string
 * \param   [in] record type character string
 * \return  t_std_error
 */

t_std_error db_sql_insert(db_sql_handle_t db_handle, const char *table_name,
    const char *record);

/**
 * \brief   Delete record(s) from the database table matching condition
 * \param   [in] db_handle type db_sql_handle_t
 * \param   [in] table_name character string
 * \param   [in] condition type character string
 * \return  t_std_error
 */

t_std_error db_sql_delete(db_sql_handle_t db_handle,
    const char *table_name, const char *condition);
/**
 * \brief   Delete all records from the database table
 * \param   [in] db_handle type db_sql_handle_t
 * \param   [in] table_name character string
 * \return  t_std_error
 */

t_std_error db_sql_delete_all_records(db_sql_handle_t db_handle,
    const char *table_name);
/**
 * \brief   Create an sql trigger
 * \param   [in] db_handle type db_sql_handle_t
 * \param   [in] sql_trigger type character string (sql trigger statement)
 * \return  t_std_error
 */

t_std_error db_sql_create_trigger(db_sql_handle_t db_handle,
    const char *sql_trigger);

/**
 * \brief   Execute sql command other than select
 * \param   [in] db_handle type db_sql_handle_t
 * \param   [in] sql_command type character string
 * \param   [in] operation_type type db_sql_opcode_t
 * \return  t_std_error
 */

t_std_error db_sql_execute_sql_command(db_sql_handle_t db_handle,
    const char *sql_command, db_sql_opcode_t operation_type);

/**
 * \brief   Execute sql select command
 * \param   [in] db_handle type db_sql_handle_t
 * \param   [in]  sql_command type character string
 * \param   [out] record_count type pointer to unsigned integer
 * \param   [out] db_ouput type array of character strings
 * \return  t_std_error
 */

t_std_error db_sql_execute_select_cmd(db_sql_handle_t db_handle,
     const char *sql_command, unsigned int *record_count, char **db_output );

/**
 * \brief   Retrieve the specific attribute from a database table
 * \param   [in] db_handle type db_sql_handle_t
 * \param   [in] table_name type character string
 * \param   [in] attribute_name type character string
 * \param   [in] value type character string
 * \param   [in] condition type character string
 * \param   [out] output_str type char string
 * \return  t_std_error
 */

t_std_error db_sql_get_attribute(db_sql_handle_t db_handle,
    const char *table_name, const char *attribute_name, const char *condition,
    char *output_str);

/**
 * \brief   Set the specific attribute in a database table
 * \param   [in] db_handle type db_sql_handle_t
 * \param   [in] table_name type character string
 * \param   [in] attribute_name type character string
 * \param   [in] value type character string
 * \param   [in] condition type character string
 * \return  t_std_error
 */

t_std_error db_sql_set_attribute(db_sql_handle_t db_handle,
    const char *table_name, const char *attribute_name, const char *value,
    const char *condition);

/**
 * \brief   Execute raw sql command
 * \param   [in] db_handle type db_sql_handle_t
 * \param   [in] input_cmd type character string
 * \param   [out] output_len type pointer to unsigned int
 * \param   [out] output_buff type character string
 * \return  t_std_error
 */

t_std_error db_sql_raw_sql_execute (db_sql_handle_t db_handle, const char *input_cmd,
    size_t *output_len, char *output_buff);

/**
 *  \}
 */

#ifdef __cplusplus
}
#endif
#endif /* _DB_SQL_OPS_H_ */
