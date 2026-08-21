<?php
/**
 * StarLoft KYC 实名认证插件配置文件
 * 
 * 此文件定义插件的配置项，系统会自动生成配置表单
 */
return [
    // ==================== 系统字段 ====================
    // 以下两个字段为系统必需字段
    
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
        'value' => 'https://kyc.starloft.cn/api/v1',
        'tip'   => 'KYC系统的API地址，例如：https://kyc.starloft.cn/api/v1',
    ],
    
    'api_key' => [
        'title' => 'API Key',
        'type'  => 'text',
        'value' => '',
        'tip'   => '在KYC系统后台获取的API Key',
    ],
    
    'api_secret' => [
        'title' => 'API Secret',
        'type'  => 'text',
        'value' => '',
        'tip'   => '在KYC系统后台获取的API Secret（用于生成HMAC签名）',
    ],
];
