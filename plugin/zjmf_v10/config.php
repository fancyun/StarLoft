<?php
/**
 * StarLoft KYC 实名认证插件配置文件（智简魔方业务系统 v10）
 *
 * 此文件定义插件的配置项，系统会自动生成配置表单（后台：实名认证 → 接口管理 → 配置）
 */
return [
    // ==================== 系统字段 ====================
    // 以下两个字段为系统必需字段（单次认证费用 / 免费认证次数）

    'amount' => [
        'title' => '单次认证费用',
        'type'  => 'text',
        'value' => '0',
        'tip'   => '每次认证扣除的费用（元），0表示不扣费',
    ],

    'free' => [
        'title' => '免费认证次数',
        'type'  => 'text',
        'value' => '0',
        'tip'   => '每个用户的免费认证次数，0表示无免费次数',
    ],

    // ==================== 插件配置字段 ====================

    'api_url' => [
        'title' => 'API地址',
        'type'  => 'text',
        'value' => 'https://www.starloft.cn/api/v1',
        'tip'   => 'StarLoft KYC系统的API地址，例如：https://www.starloft.cn/api/v1',
    ],

    'api_key' => [
        'title' => 'API Key',
        'type'  => 'text',
        'value' => '',
        'tip'   => '在KYC系统后台「用户中心 → API密钥管理」获取',
    ],

    'api_secret' => [
        'title' => 'API Secret',
        'type'  => 'text',
        'value' => '',
        'tip'   => '在KYC系统后台获取（用于生成HMAC签名），请妥善保管',
    ],

    'min_age' => [
        'title' => '最低实名年龄',
        'type'  => 'text',
        'value' => '16',
        'tip'   => '实名认证要求的最低年龄（周岁），根据身份证号中的出生日期计算，0表示不限年龄',
    ],

    'return_url' => [
        'title' => '认证完成回跳地址',
        'type'  => 'text',
        'value' => '',
        'tip'   => '用户完成认证后浏览器回跳地址；留空则使用插件内置结果页',
    ],
];
